# tidb grafana监控发现大量raft group缺少副本问题跟踪

近日tidb集群抖动严重，如果grafana的监控发现，有大量的raft group出现缺少副本的情况。就是下图显示的`Region Health`中最多的时候有2万7千个region出现`miss_peer_region_count`

[]

## miss_peer_region_count 出处


### 如何监控

```go
// server/cluster.go
func (c *RaftCluster) runBackgroundJobs(interval time.Duration) {
	defer logutil.LogPanic()
	defer c.wg.Done()
	
	// interval 是一分钟，使用的是：var backgroundJobInterval = time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.quit:
			log.Info("background jobs has been stopped")
			return
		case <-ticker.C:
			c.checkOperators()
			c.checkStores()
			c.collectMetrics()
			c.coordinator.pruneHistory()
		}
	}
}

func (c *RaftCluster) collectMetrics() {
	cluster := c.cachedCluster
	statsMap := newStoreStatisticsMap(c.cachedCluster.opt, c.GetNamespaceClassifier())
	for _, s := range cluster.GetStores() {
		statsMap.Observe(s)
	}
	statsMap.Collect()

	c.coordinator.collectSchedulerMetrics()
	c.coordinator.collectHotSpotMetrics()
	cluster.collectMetrics()
	c.collectHealthStatus()
}
```

```go
// server/cluster_info.go
func (c *clusterInfo) collectMetrics() {
	if c.regionStats == nil {
		return
	}
	c.RLock()
	defer c.RUnlock()
	c.regionStats.Collect()
	c.labelLevelStats.Collect()
	// collect hot cache metrics
	c.core.HotCache.CollectMetrics(c.core.Stores)
}
```

```go
// server/region_statistics.go
func (r *regionStatistics) Collect() {
	regionStatusGauge.WithLabelValues("miss_peer_region_count").Set(float64(len(r.stats[missPeer])))
	regionStatusGauge.WithLabelValues("extra_peer_region_count").Set(float64(len(r.stats[extraPeer])))
	regionStatusGauge.WithLabelValues("down_peer_region_count").Set(float64(len(r.stats[downPeer])))
	regionStatusGauge.WithLabelValues("pending_peer_region_count").Set(float64(len(r.stats[pendingPeer])))
	regionStatusGauge.WithLabelValues("offline_peer_region_count").Set(float64(len(r.stats[offlinePeer])))
	regionStatusGauge.WithLabelValues("incorrect_namespace_region_count").Set(float64(len(r.stats[incorrectNamespace])))
	regionStatusGauge.WithLabelValues("learner_peer_region_count").Set(float64(len(r.stats[learnerPeer])))
}
```

### 具体数据来源

```go
// server/grpc_service.go
// RegionHeartbeat implements gRPC PDServer.
func (s *Server) RegionHeartbeat(stream pdpb.PD_RegionHeartbeatServer) error {
	server := &heartbeatServer{stream: stream}
	cluster := s.GetRaftCluster()
	if cluster == nil {
		resp := &pdpb.RegionHeartbeatResponse{
			Header: s.notBootstrappedHeader(),
		}
		err := server.Send(resp)
		return errors.WithStack(err)
	}

	var lastBind time.Time
	for {
		request, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.WithStack(err)
		}

		if err = s.validateRequest(request.GetHeader()); err != nil {
			return err
		}

		storeID := request.GetLeader().GetStoreId()
		storeLabel := strconv.FormatUint(storeID, 10)

		regionHeartbeatCounter.WithLabelValues(storeLabel, "report", "recv").Inc()
		regionHeartbeatLatency.WithLabelValues(storeLabel).Observe(float64(time.Now().Unix()) - float64(request.GetInterval().GetEndTimestamp()))

		hbStreams := cluster.coordinator.hbStreams

		if time.Since(lastBind) > s.cfg.heartbeatStreamBindInterval.Duration {
			regionHeartbeatCounter.WithLabelValues(storeLabel, "report", "bind").Inc()
			hbStreams.bindStream(storeID, server)
			lastBind = time.Now()
		}

		region := core.RegionFromHeartbeat(request)
		if region.GetID() == 0 {
			msg := fmt.Sprintf("invalid request region, %v", request)
			hbStreams.sendErr(region, pdpb.ErrorType_UNKNOWN, msg, storeLabel)
			continue
		}
		if region.GetLeader() == nil {
			msg := fmt.Sprintf("invalid request leader, %v", request)
			hbStreams.sendErr(region, pdpb.ErrorType_UNKNOWN, msg, storeLabel)
			continue
		}

		err = cluster.HandleRegionHeartbeat(region)
		if err != nil {
			msg := err.Error()
			hbStreams.sendErr(region, pdpb.ErrorType_UNKNOWN, msg, storeLabel)
		}

		regionHeartbeatCounter.WithLabelValues(storeLabel, "report", "ok").Inc()
	}
}
```
```go
// server/cluster_worker.go
// HandleRegionHeartbeat processes RegionInfo reports from client.
func (c *RaftCluster) HandleRegionHeartbeat(region *core.RegionInfo) error {
	if err := c.cachedCluster.handleRegionHeartbeat(region); err != nil {
		return err
	}

	// If the region peer count is 0, then we should not handle this.
	if len(region.GetPeers()) == 0 {
		log.Warn("invalid region, zero region peer count", zap.Stringer("region-meta", core.RegionToHexMeta(region.GetMeta())))
		return errors.Errorf("invalid region, zero region peer count: %v", core.RegionToHexMeta(region.GetMeta()))
	}

	c.coordinator.dispatch(region)
	return nil
}
```

```go
// server/cluster_info.go

// handleRegionHeartbeat updates the region information.
func (c *clusterInfo) handleRegionHeartbeat(region *core.RegionInfo) error {
	c.RLock()
	origin := c.core.Regions.GetRegion(region.GetID())
	if origin == nil {
		for _, item := range c.core.Regions.GetOverlaps(region) {
			if region.GetRegionEpoch().GetVersion() < item.GetRegionEpoch().GetVersion() {
				c.RUnlock()
				return ErrRegionIsStale(region.GetMeta(), item)
			}
		}
	}
	isWriteUpdate, writeItem := c.core.CheckWriteStatus(region)
	isReadUpdate, readItem := c.core.CheckReadStatus(region)
	c.RUnlock()

	// Save to KV if meta is updated.
	// Save to cache if meta or leader is updated, or contains any down/pending peer.
	// Mark isNew if the region in cache does not have leader.
	var saveKV, saveCache, isNew bool
	if origin == nil {
		log.Debug("insert new region",
			zap.Uint64("region-id", region.GetID()),
			zap.Stringer("meta-region", core.RegionToHexMeta(region.GetMeta())),
		)
		saveKV, saveCache, isNew = true, true, true
	} else {
		r := region.GetRegionEpoch()
		o := origin.GetRegionEpoch()
		// Region meta is stale, return an error.
		if r.GetVersion() < o.GetVersion() || r.GetConfVer() < o.GetConfVer() {
			return ErrRegionIsStale(region.GetMeta(), origin.GetMeta())
		}
		if r.GetVersion() > o.GetVersion() {
			log.Info("region Version changed",
				zap.Uint64("region-id", region.GetID()),
				zap.String("detail", core.DiffRegionKeyInfo(origin, region)),
				zap.Uint64("old-version", o.GetVersion()),
				zap.Uint64("new-version", r.GetVersion()),
			)
			saveKV, saveCache = true, true
		}
		if r.GetConfVer() > o.GetConfVer() {
			log.Info("region ConfVer changed",
				zap.Uint64("region-id", region.GetID()),
				zap.String("detail", core.DiffRegionPeersInfo(origin, region)),
				zap.Uint64("old-confver", o.GetConfVer()),
				zap.Uint64("new-confver", r.GetConfVer()),
			)
			saveKV, saveCache = true, true
		}
		if region.GetLeader().GetId() != origin.GetLeader().GetId() {
			if origin.GetLeader().GetId() == 0 {
				isNew = true
			} else {
				log.Info("leader changed",
					zap.Uint64("region-id", region.GetID()),
					zap.Uint64("from", origin.GetLeader().GetStoreId()),
					zap.Uint64("to", region.GetLeader().GetStoreId()),
				)
			}
			saveCache = true
		}
		if len(region.GetDownPeers()) > 0 || len(region.GetPendingPeers()) > 0 {
			saveCache = true
		}
		if len(origin.GetDownPeers()) > 0 || len(origin.GetPendingPeers()) > 0 {
			saveCache = true
		}
		if len(region.GetPeers()) != len(origin.GetPeers()) {
			saveKV, saveCache = true, true
		}
		if region.GetApproximateSize() != origin.GetApproximateSize() {
			saveCache = true
		}
		if region.GetApproximateKeys() != origin.GetApproximateKeys() {
			saveCache = true
		}
	}

	if saveKV && c.kv != nil {
		if err := c.kv.SaveRegion(region.GetMeta()); err != nil {
			// Not successfully saved to kv is not fatal, it only leads to longer warm-up
			// after restart. Here we only log the error then go on updating cache.
			log.Error("fail to save region to kv",
				zap.Uint64("region-id", region.GetID()),
				zap.Stringer("region-meta", core.RegionToHexMeta(region.GetMeta())),
				zap.Error(err))
		}
	}
	if !isWriteUpdate && !isReadUpdate && !saveCache && !isNew {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	if isNew {
		c.prepareChecker.collect(region)
	}

	if saveCache {
		overlaps := c.core.Regions.SetRegion(region)
		if c.kv != nil {
			for _, item := range overlaps {
				if err := c.kv.DeleteRegion(item); err != nil {
					log.Error("fail to delete region from kv",
						zap.Uint64("region-id", item.GetId()),
						zap.Stringer("region-meta", core.RegionToHexMeta(item)),
						zap.Error(err))
				}
			}
		}
		for _, item := range overlaps {
			if c.regionStats != nil {
				c.regionStats.clearDefunctRegion(item.GetId())
			}
			c.labelLevelStats.clearDefunctRegion(item.GetId())
		}

		// Update related stores.
		if origin != nil {
			for _, p := range origin.GetPeers() {
				c.updateStoreStatusLocked(p.GetStoreId())
			}
		}
		for _, p := range region.GetPeers() {
			c.updateStoreStatusLocked(p.GetStoreId())
		}
	}

	if c.regionStats != nil {
		c.regionStats.Observe(region, c.takeRegionStoresLocked(region))
	}

	key := region.GetID()
	if isWriteUpdate {
		c.core.HotCache.Update(key, writeItem, schedule.WriteFlow)
	}
	if isReadUpdate {
		c.core.HotCache.Update(key, readItem, schedule.ReadFlow)
	}
	return nil
}
```

```go
// server/region_statistics.go

func (r *regionStatistics) Observe(region *core.RegionInfo, stores []*core.StoreInfo) {
	// Region state.
	regionID := region.GetID()
	namespace := r.classifier.GetRegionNamespace(region)
	var (
		peerTypeIndex regionStatisticType
		deleteIndex   regionStatisticType
	)
	if len(region.GetPeers()) < r.opt.GetMaxReplicas(namespace) {
		r.stats[missPeer][regionID] = region
		peerTypeIndex |= missPeer
	} else if len(region.GetPeers()) > r.opt.GetMaxReplicas(namespace) {
		r.stats[extraPeer][regionID] = region
		peerTypeIndex |= extraPeer
	}

	if len(region.GetDownPeers()) > 0 {
		r.stats[downPeer][regionID] = region
		peerTypeIndex |= downPeer
	}

	if len(region.GetPendingPeers()) > 0 {
		r.stats[pendingPeer][regionID] = region
		peerTypeIndex |= pendingPeer
	}

	if len(region.GetLearners()) > 0 {
		r.stats[learnerPeer][regionID] = region
		peerTypeIndex |= learnerPeer
	}

	for _, store := range stores {
		if store.IsOffline() {
			peer := region.GetStorePeer(store.GetId())
			if peer != nil {
				r.stats[offlinePeer][regionID] = region
				peerTypeIndex |= offlinePeer
			}
		}
		ns := r.classifier.GetStoreNamespace(store)
		if ns == namespace {
			continue
		}
		r.stats[incorrectNamespace][regionID] = region
		peerTypeIndex |= incorrectNamespace
		break
	}

	if oldIndex, ok := r.index[regionID]; ok {
		deleteIndex = oldIndex &^ peerTypeIndex
	}
	r.deleteEntry(deleteIndex, regionID)
	r.index[regionID] = peerTypeIndex
}
```


https://github.com/pingcap/tidb-ansible/blob/master/roles/prometheus/files/pd.rules.yml

```
  - alert: PD_miss_peer_region_count
    expr: sum( pd_regions_status{type="miss_peer_region_count"} )  > 100
    for: 1m
    labels:
      env: ENV_LABELS_ENV
      level: critical
      expr:  sum( pd_regions_status{type="miss_peer_region_count"} )  > 100
    annotations:
      description: 'cluster: ENV_LABELS_ENV, instance: {{ $labels.instance }}, values:{{ $value }}'
      value: '{{ $value }}'
      summary: PD_miss_peer_region_count
```

##  如何补peer的


```go
// server/cluster.go
func (c *RaftCluster) runCoordinator() {
	defer logutil.LogPanic()
	defer c.wg.Done()
	defer func() {
		c.coordinator.wg.Wait()
		log.Info("coordinator has been stopped")
	}()
	c.coordinator.run()
	<-c.coordinator.ctx.Done()
	log.Info("coordinator is stopping")
}

```

```go
// server/coordiantor.go

func (c *coordinator) run() {
	ticker := time.NewTicker(runSchedulerCheckInterval)
	defer ticker.Stop()
	log.Info("coordinator starts to collect cluster information")
	for {
		if c.shouldRun() {
			log.Info("coordinator has finished cluster information preparation")
			break
		}
		select {
		case <-ticker.C:
		case <-c.ctx.Done():
			log.Info("coordinator stops running")
			return
		}
	}
	log.Info("coordinator starts to run schedulers")

	k := 0
	scheduleCfg := c.cluster.opt.load()
	for _, schedulerCfg := range scheduleCfg.Schedulers {
		if schedulerCfg.Disable {
			scheduleCfg.Schedulers[k] = schedulerCfg
			k++
			log.Info("skip create scheduler", zap.String("scheduler-type", schedulerCfg.Type))
			continue
		}
		s, err := schedule.CreateScheduler(schedulerCfg.Type, c.limiter, schedulerCfg.Args...)
		if err != nil {
			log.Fatal("can not create scheduler", zap.String("scheduler-type", schedulerCfg.Type), zap.Error(err))
		}
		log.Info("create scheduler", zap.String("scheduler-name", s.GetName()))
		if err = c.addScheduler(s, schedulerCfg.Args...); err != nil {
			log.Error("can not add scheduler", zap.String("scheduler-name", s.GetName()), zap.Error(err))
		}

		// only record valid scheduler config
		if err == nil {
			scheduleCfg.Schedulers[k] = schedulerCfg
			k++
		}
	}

	// remove invalid scheduler config and persist
	scheduleCfg.Schedulers = scheduleCfg.Schedulers[:k]
	if err := c.cluster.opt.persist(c.cluster.kv); err != nil {
		log.Error("cannot persist schedule config", zap.Error(err))
	}

	c.wg.Add(1)
	go c.patrolRegions()
}


func (c *coordinator) patrolRegions() {
	defer logutil.LogPanic()

	defer c.wg.Done()
	timer := time.NewTimer(c.cluster.GetPatrolRegionInterval())
	defer timer.Stop()

	log.Info("coordinator starts patrol regions")
	start := time.Now()
	var key []byte
	for {
		select {
		case <-timer.C:
			timer.Reset(c.cluster.GetPatrolRegionInterval())
		case <-c.ctx.Done():
			log.Info("patrol regions has been stopped")
			return
		}

		regions := c.cluster.ScanRegions(key, patrolScanRegionLimit)
		if len(regions) == 0 {
			// reset scan key.
			key = nil
			continue
		}

		for _, region := range regions {
			// Skip the region if there is already a pending operator.
			if c.getOperator(region.GetID()) != nil {
				continue
			}

			key = region.GetEndKey()

			if c.checkRegion(region) {
				break
			}
		}
		// update label level isolation statistics.
		c.cluster.updateRegionsLabelLevelStats(regions)
		if len(key) == 0 {
			patrolCheckRegionsHistogram.Observe(time.Since(start).Seconds())
			start = time.Now()
		}
	}
}



func (c *coordinator) checkRegion(region *core.RegionInfo) bool {
	// If PD has restarted, it need to check learners added before and promote them.
	// Don't check isRaftLearnerEnabled cause it may be disable learner feature but still some learners to promote.
	for _, p := range region.GetLearners() {
		if region.GetPendingLearner(p.GetId()) != nil {
			continue
		}
		step := schedule.PromoteLearner{
			ToStore: p.GetStoreId(),
			PeerID:  p.GetId(),
		}
		op := schedule.NewOperator("promoteLearner", region.GetID(), region.GetRegionEpoch(), schedule.OpRegion, step)
		if c.addOperator(op) {
			return true
		}
	}

	if c.limiter.OperatorCount(schedule.OpLeader) < c.cluster.GetLeaderScheduleLimit() &&
		c.limiter.OperatorCount(schedule.OpRegion) < c.cluster.GetRegionScheduleLimit() &&
		c.limiter.OperatorCount(schedule.OpReplica) < c.cluster.GetReplicaScheduleLimit() {
		if op := c.namespaceChecker.Check(region); op != nil {
			if c.addOperator(op) {
				return true
			}
		}
	}

	if c.limiter.OperatorCount(schedule.OpReplica) < c.cluster.GetReplicaScheduleLimit() {
		if op := c.replicaChecker.Check(region); op != nil {
			if c.addOperator(op) {
				return true
			}
		}
	}
	if c.cluster.IsFeatureSupported(RegionMerge) && c.limiter.OperatorCount(schedule.OpMerge) < c.cluster.GetMergeScheduleLimit() {
		if op1, op2 := c.mergeChecker.Check(region); op1 != nil && op2 != nil {
			// make sure two operators can add successfully altogether
			if c.addOperator(op1, op2) {
				return true
			}
		}
	}
	return false
}


```

```go
// server/scheduler/replica_checker.go

// Check verifies a region's replicas, creating an Operator if need.
func (r *ReplicaChecker) Check(region *core.RegionInfo) *Operator {
	checkerCounter.WithLabelValues("replica_checker", "check").Inc()
	if op := r.checkDownPeer(region); op != nil {
		checkerCounter.WithLabelValues("replica_checker", "new_operator").Inc()
		op.SetPriorityLevel(core.HighPriority)
		return op
	}
	if op := r.checkOfflinePeer(region); op != nil {
		checkerCounter.WithLabelValues("replica_checker", "new_operator").Inc()
		op.SetPriorityLevel(core.HighPriority)
		return op
	}

	if len(region.GetPeers()) < r.cluster.GetMaxReplicas() && r.cluster.IsMakeUpReplicaEnabled() {
		log.Debug("region has fewer than max replicas", zap.Uint64("region-id", region.GetID()), zap.Int("peers", len(region.GetPeers())))
		newPeer, _ := r.selectBestPeerToAddReplica(region, NewStorageThresholdFilter())
		if newPeer == nil {
			checkerCounter.WithLabelValues("replica_checker", "no_target_store").Inc()
			return nil
		}
		var steps []OperatorStep
		if r.cluster.IsRaftLearnerEnabled() {
			steps = []OperatorStep{
				AddLearner{ToStore: newPeer.GetStoreId(), PeerID: newPeer.GetId()},
				PromoteLearner{ToStore: newPeer.GetStoreId(), PeerID: newPeer.GetId()},
			}
		} else {
			steps = []OperatorStep{
				AddPeer{ToStore: newPeer.GetStoreId(), PeerID: newPeer.GetId()},
			}
		}
		checkerCounter.WithLabelValues("replica_checker", "new_operator").Inc()
		return NewOperator("makeUpReplica", region.GetID(), region.GetRegionEpoch(), OpReplica|OpRegion, steps...)
	}

	// when add learner peer, the number of peer will exceed max replicas for a while,
	// just comparing the the number of voters to avoid too many cancel add operator log.
	if len(region.GetVoters()) > r.cluster.GetMaxReplicas() && r.cluster.IsRemoveExtraReplicaEnabled() {
		log.Debug("region has more than max replicas", zap.Uint64("region-id", region.GetID()), zap.Int("peers", len(region.GetPeers())))
		oldPeer, _ := r.selectWorstPeer(region)
		if oldPeer == nil {
			checkerCounter.WithLabelValues("replica_checker", "no_worst_peer").Inc()
			return nil
		}
		op, err := CreateRemovePeerOperator("removeExtraReplica", r.cluster, OpReplica, region, oldPeer.GetStoreId())
		if err != nil {
			checkerCounter.WithLabelValues("replica_checker", "create_operator_fail").Inc()
			return nil
		}
		checkerCounter.WithLabelValues("replica_checker", "new_operator").Inc()
		return op
	}

	return r.checkBestReplacement(region)
}
```




