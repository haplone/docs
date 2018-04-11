
# 使用cobra来处理系统启动的命令行交互

* main.go中import 	"github.com/cockroachdb/cockroach/pkg/cli"
* 在cli/cli.go中使用init方法对md进行初始化

```go
// cli/cli.go
func init() {
	// ...

	cockroachCmd.AddCommand(
		StartCmd,
		initCmd,
		certCmd,
		quitCmd,

		sqlShellCmd,
		userCmd,
		zoneCmd,
		nodeCmd,
		dumpCmd,

		// Miscellaneous commands.
		genCmd,
		versionCmd,
		debugCmd,
	)
}
```
* 在cli/start.go中有StartCmd的定义

```go
// cli/start.go
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a node",
	Long: `
Start a CockroachDB node, which will export data from one or more
storage devices, specified via --store flags.

If no cluster exists yet and this is the first node, no additional
flags are required. If the cluster already exists, and this node is
uninitialized, specify the --join flag to point to any healthy node
(or list of nodes) already part of the cluster.
`,
	Example: `  cockroach start --insecure --store=attrs=ssd,path=/mnt/ssd1 [--join=host:port,[host:port]]`,
	RunE:    MaybeShoutError(MaybeDecorateGRPCError(runStart)),
}
```
* 然后系统才开始通过runStart接手进行系统初始化

# 启动server

* 一堆的参数检查设置、log准备
* new 一个server
* 启动server

```go
// cli/start.go
func runStart(cmd *cobra.Command, args []string) error {
	// ...

	// 使用分布式追踪协议opentracing
	tracer := serverCfg.Settings.Tracer
	sp := tracer.StartSpan("server start")
	ctx := opentracing.ContextWithSpan(context.Background(), sp)

	// Set up the logging and profiling output.
	stopper, err := setupAndInitializeLoggingAndProfiling(ctx)
  // ...
	// grpc 出场
	grpcutil.SetSeverity(log.Severity_WARNING)

	// ...
	log.Info(ctx, "starting cockroach node")

	// ...
	var s *server.Server
	errChan := make(chan error, 1)
	go func() {
		// ...
		if err := func() error {
			// Instantiate the server.
			var err error
			s, err = server.NewServer(serverCfg, stopper)
			// ...

			// Attempt to start the server.
			if err := s.Start(ctx); err != nil {
				// ...
			}
			// ...
			// 一堆的善后处理
	}()
	// ... 应用关闭的处理

	return returnErr
}
```

# new server

* 初始化Server
```go
// server/server.go
// NewServer creates a Server from a server.Config.
func NewServer(cfg Config, stopper *stop.Stopper) (*Server, error) {
	// ...
	s := &Server{
		st:       st,
		mux:      http.NewServeMux(),
		clock:    hlc.NewClock(hlc.UnixNano, time.Duration(cfg.MaxOffset)),
		stopper:  stopper,
		cfg:      cfg,
		registry: metric.NewRegistry(),
	}

	// ...

	s.grpc = rpc.NewServerWithInterceptor(s.rpcContext, s.Intercept())

	s.gossip = gossip.New(
		s.cfg.AmbientCtx,
		&s.rpcContext.ClusterID,
		&s.nodeIDContainer,
		s.rpcContext,
		s.grpc,
		s.stopper,
		s.registry,
	)

	// A custom RetryOptions is created which uses stopper.ShouldQuiesce() as
	// the Closer. This prevents infinite retry loops from occurring during
	// graceful server shutdown
	//
	// Such a loop occurs when the DistSender attempts a connection to the
	// local server during shutdown, and receives an internal server error (HTTP
	// Code 5xx). This is the correct error for a server to return when it is
	// shutting down, and is normally retryable in a cluster environment.
	// However, on a single-node setup (such as a test), retries will never
	// succeed because the only server has been shut down; thus, the
	// DistSender needs to know that it should not retry in this situation.
	retryOpts := s.cfg.RetryOptions
	if retryOpts == (retry.Options{}) {
		retryOpts = base.DefaultRetryOptions()
	}
	retryOpts.Closer = s.stopper.ShouldQuiesce()
	distSenderCfg := kv.DistSenderConfig{
		AmbientCtx:      s.cfg.AmbientCtx,
		Settings:        st,
		Clock:           s.clock,
		RPCContext:      s.rpcContext,
		RPCRetryOptions: &retryOpts,
	}
	if distSenderTestingKnobs := s.cfg.TestingKnobs.DistSender; distSenderTestingKnobs != nil {
		distSenderCfg.TestingKnobs = *distSenderTestingKnobs.(*kv.DistSenderTestingKnobs)
	}
	s.distSender = kv.NewDistSender(distSenderCfg, s.gossip)
	s.registry.AddMetricStruct(s.distSender.Metrics())

	txnMetrics := kv.MakeTxnMetrics(s.cfg.HistogramWindowInterval())
	s.registry.AddMetricStruct(txnMetrics)
	s.tcsFactory = kv.NewTxnCoordSenderFactory(
		s.cfg.AmbientCtx,
		st,
		s.distSender,
		s.clock,
		s.cfg.Linearizable,
		s.stopper,
		txnMetrics,
	)
	dbCtx := client.DefaultDBContext()
	dbCtx.NodeID = &s.nodeIDContainer
	s.db = client.NewDBWithContext(s.tcsFactory, s.clock, dbCtx)

	nlActive, nlRenewal := s.cfg.NodeLivenessDurations()

	s.nodeLiveness = storage.NewNodeLiveness(
		s.cfg.AmbientCtx,
		s.clock,
		s.db,
		s.gossip,
		nlActive,
		nlRenewal,
		s.cfg.HistogramWindowInterval(),
	)
	s.registry.AddMetricStruct(s.nodeLiveness.Metrics())

	s.storePool = storage.NewStorePool(
		s.cfg.AmbientCtx,
		s.st,
		s.gossip,
		s.clock,
		storage.MakeStorePoolNodeLivenessFunc(s.nodeLiveness),
		/* deterministic */ false,
	)

	s.raftTransport = storage.NewRaftTransport(
		s.cfg.AmbientCtx, st, storage.GossipAddressResolver(s.gossip), s.grpc, s.rpcContext,
	)

	// Set up internal memory metrics for use by internal SQL executors.
	s.internalMemMetrics = sql.MakeMemMetrics("internal", cfg.HistogramWindowInterval())
	s.registry.AddMetricStruct(s.internalMemMetrics)

	// Set up Lease Manager
	var lmKnobs sql.LeaseManagerTestingKnobs
	if leaseManagerTestingKnobs := cfg.TestingKnobs.SQLLeaseManager; leaseManagerTestingKnobs != nil {
		lmKnobs = *leaseManagerTestingKnobs.(*sql.LeaseManagerTestingKnobs)
	}
	s.leaseMgr = sql.NewLeaseManager(
		s.cfg.AmbientCtx,
		nil, /* execCfg - will be set later because of circular dependencies */
		lmKnobs,
		s.stopper,
		&s.internalMemMetrics,
		s.cfg.LeaseManagerConfig,
	)

	// We do not set memory monitors or a noteworthy limit because the children of
	// this monitor will be setting their own noteworthy limits.
	rootSQLMemoryMonitor := mon.MakeMonitor(
		"root",
		mon.MemoryResource,
		nil,           /* curCount */
		nil,           /* maxHist */
		-1,            /* increment: use default increment */
		math.MaxInt64, /* noteworthy */
		st,
	)
	rootSQLMemoryMonitor.Start(context.Background(), nil, mon.MakeStandaloneBudget(s.cfg.SQLMemoryPoolSize))

	// Set up the DistSQL temp engine.

	tempEngine, err := engine.NewTempEngine(s.cfg.TempStorageConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp storage")
	}
	s.stopper.AddCloser(tempEngine)
	// Remove temporary directory linked to tempEngine after closing
	// tempEngine.
	s.stopper.AddCloser(stop.CloserFn(func() {
		firstStore := cfg.Stores.Specs[0]
		var err error
		if firstStore.InMemory {
			// First store is in-memory so we remove the temp
			// directory directly since there is no record file.
			err = os.RemoveAll(s.cfg.TempStorageConfig.Path)
		} else {
			// If record file exists, we invoke CleanupTempDirs to
			// also remove the record after the temp directory is
			// removed.
			recordPath := filepath.Join(firstStore.Path, TempDirsRecordFilename)
			err = engine.CleanupTempDirs(recordPath)
		}
		if err != nil {
			log.Errorf(context.TODO(), "could not remove temporary store directory: %v", err.Error())
		}
	}))

	// Set up admin memory metrics for use by admin SQL executors.
	s.adminMemMetrics = sql.MakeMemMetrics("admin", cfg.HistogramWindowInterval())
	s.registry.AddMetricStruct(s.adminMemMetrics)

	s.tsDB = ts.NewDB(s.db, s.cfg.Settings)
	s.registry.AddMetricStruct(s.tsDB.Metrics())
	nodeCountFn := func() int64 {
		return s.nodeLiveness.Metrics().LiveNodes.Value()
	}
	s.tsServer = ts.MakeServer(s.cfg.AmbientCtx, s.tsDB, nodeCountFn, s.cfg.TimeSeriesServerConfig, s.stopper)

	// The InternalExecutor will be further initialized later, as we create more
	// of the server's components. There's a circular dependency - many things
	// need an InternalExecutor, but the InternalExecutor needs an ExecutorConfig,
	// which in turn needs many things. That's why everybody that needs an
	// InternalExecutor takes pointers to this one instance.
	sqlExecutor := sql.InternalExecutor{}

	// Similarly for execCfg.
	var execCfg sql.ExecutorConfig

	// TODO(bdarnell): make StoreConfig configurable.
	storeCfg := storage.StoreConfig{
		Settings:                st,
		AmbientCtx:              s.cfg.AmbientCtx,
		RaftConfig:              s.cfg.RaftConfig,
		Clock:                   s.clock,
		DB:                      s.db,
		Gossip:                  s.gossip,
		NodeLiveness:            s.nodeLiveness,
		Transport:               s.raftTransport,
		RPCContext:              s.rpcContext,
		ScanInterval:            s.cfg.ScanInterval,
		ScanMaxIdleTime:         s.cfg.ScanMaxIdleTime,
		TimestampCachePageSize:  s.cfg.TimestampCachePageSize,
		HistogramWindowInterval: s.cfg.HistogramWindowInterval(),
		StorePool:               s.storePool,
		SQLExecutor:             &sqlExecutor,
		LogRangeEvents:          s.cfg.EventLogEnabled,
		TimeSeriesDataStore:     s.tsDB,

		EnableEpochRangeLeases: true,
	}
	if storeTestingKnobs := s.cfg.TestingKnobs.Store; storeTestingKnobs != nil {
		storeCfg.TestingKnobs = *storeTestingKnobs.(*storage.StoreTestingKnobs)
	}

	s.recorder = status.NewMetricsRecorder(s.clock, s.nodeLiveness, s.rpcContext, s.gossip, st)
	s.registry.AddMetricStruct(s.rpcContext.RemoteClocks.Metrics())

	s.runtime = status.MakeRuntimeStatSampler(s.clock)
	s.registry.AddMetricStruct(s.runtime)

	s.node = NewNode(
		storeCfg, s.recorder, s.registry, s.stopper,
		txnMetrics, nil /* execCfg */, &s.rpcContext.ClusterID)
	roachpb.RegisterInternalServer(s.grpc, s.node)
	storage.RegisterConsistencyServer(s.grpc, s.node.storesServer)

	s.sessionRegistry = sql.MakeSessionRegistry()
	s.jobRegistry = jobs.MakeRegistry(
		s.cfg.AmbientCtx, s.clock, s.db, &sqlExecutor, &s.nodeIDContainer, st, func(opName, user string) (interface{}, func()) {
			// This is a hack to get around a Go package dependency cycle. See comment
			// in sql/jobs/registry.go on planHookMaker.
			return sql.NewInternalPlanner(opName, nil, user, &sql.MemoryMetrics{}, &execCfg)
		})

	distSQLMetrics := distsqlrun.MakeDistSQLMetrics(cfg.HistogramWindowInterval())
	s.registry.AddMetricStruct(distSQLMetrics)

	// Set up the DistSQL server.
	distSQLCfg := distsqlrun.ServerConfig{
		AmbientContext: s.cfg.AmbientCtx,
		Settings:       st,
		DB:             s.db,
		Executor:       &sqlExecutor,
		FlowDB:         client.NewDB(s.tcsFactory, s.clock),
		RPCContext:     s.rpcContext,
		Stopper:        s.stopper,
		NodeID:         &s.nodeIDContainer,

		TempStorage: tempEngine,
		DiskMonitor: s.cfg.TempStorageConfig.Mon,

		ParentMemoryMonitor: &rootSQLMemoryMonitor,

		Metrics: &distSQLMetrics,

		JobRegistry: s.jobRegistry,
		Gossip:      s.gossip,
	}
	if distSQLTestingKnobs := s.cfg.TestingKnobs.DistSQL; distSQLTestingKnobs != nil {
		distSQLCfg.TestingKnobs = *distSQLTestingKnobs.(*distsqlrun.TestingKnobs)
	}

	s.distSQLServer = distsqlrun.NewServer(ctx, distSQLCfg)
	distsqlrun.RegisterDistSQLServer(s.grpc, s.distSQLServer)

	s.admin = newAdminServer(s, &sqlExecutor)
	s.status = newStatusServer(
		s.cfg.AmbientCtx,
		st,
		s.cfg.Config,
		s.admin,
		s.db,
		s.gossip,
		s.recorder,
		s.nodeLiveness,
		s.rpcContext,
		s.node.stores,
		s.stopper,
		s.sessionRegistry,
	)
	s.authentication = newAuthenticationServer(s, &sqlExecutor)
	for _, gw := range []grpcGatewayServer{s.admin, s.status, s.authentication, &s.tsServer} {
		gw.RegisterService(s.grpc)
	}

	s.initServer = newInitServer(s)
	s.initServer.semaphore.acquire()

	serverpb.RegisterInitServer(s.grpc, s.initServer)

	nodeInfo := sql.NodeInfo{
		AdminURL:  cfg.AdminURL,
		PGURL:     cfg.PGURL,
		ClusterID: s.ClusterID,
		NodeID:    &s.nodeIDContainer,
	}

	virtualSchemas, err := sql.NewVirtualSchemaHolder(ctx, st)
	if err != nil {
		log.Fatal(ctx, err)
	}

	// Set up Executor

	var sqlExecutorTestingKnobs *sql.ExecutorTestingKnobs
	if k := s.cfg.TestingKnobs.SQLExecutor; k != nil {
		sqlExecutorTestingKnobs = k.(*sql.ExecutorTestingKnobs)
	} else {
		sqlExecutorTestingKnobs = new(sql.ExecutorTestingKnobs)
	}

	execCfg = sql.ExecutorConfig{
		Settings:                s.st,
		NodeInfo:                nodeInfo,
		AmbientCtx:              s.cfg.AmbientCtx,
		DB:                      s.db,
		Gossip:                  s.gossip,
		DistSender:              s.distSender,
		RPCContext:              s.rpcContext,
		LeaseManager:            s.leaseMgr,
		Clock:                   s.clock,
		DistSQLSrv:              s.distSQLServer,
		StatusServer:            s.status,
		SessionRegistry:         s.sessionRegistry,
		JobRegistry:             s.jobRegistry,
		VirtualSchemas:          virtualSchemas,
		HistogramWindowInterval: s.cfg.HistogramWindowInterval(),
		RangeDescriptorCache:    s.distSender.RangeDescriptorCache(),
		LeaseHolderCache:        s.distSender.LeaseHolderCache(),
		TestingKnobs:            sqlExecutorTestingKnobs,
		DistSQLPlanner: sql.NewDistSQLPlanner(
			ctx,
			distsqlrun.Version,
			s.st,
			// The node descriptor will be set later, once it is initialized.
			roachpb.NodeDescriptor{},
			s.rpcContext,
			s.distSQLServer,
			s.distSender,
			s.gossip,
			s.stopper,
			sqlExecutorTestingKnobs.DistSQLPlannerKnobs,
		),
		ExecLogger:             log.NewSecondaryLogger(nil, "sql-exec", true /*enableGc*/, false /*forceSyncWrites*/),
		AuditLogger:            log.NewSecondaryLogger(s.cfg.SQLAuditLogDirName, "sql-audit", true /*enableGc*/, true /*forceSyncWrites*/),
		ConnResultsBufferBytes: s.cfg.ConnResultsBufferBytes,
	}

	if sqlSchemaChangerTestingKnobs := s.cfg.TestingKnobs.SQLSchemaChanger; sqlSchemaChangerTestingKnobs != nil {
		execCfg.SchemaChangerTestingKnobs = sqlSchemaChangerTestingKnobs.(*sql.SchemaChangerTestingKnobs)
	} else {
		execCfg.SchemaChangerTestingKnobs = new(sql.SchemaChangerTestingKnobs)
	}
	if sqlEvalContext := s.cfg.TestingKnobs.SQLEvalContext; sqlEvalContext != nil {
		execCfg.EvalContextTestingKnobs = *sqlEvalContext.(*tree.EvalContextTestingKnobs)
	}
	s.sqlExecutor = sql.NewExecutor(execCfg, s.stopper)
	if s.cfg.UseLegacyConnHandling {
		s.registry.AddMetricStruct(s.sqlExecutor)
	}

	s.pgServer = pgwire.MakeServer(
		s.cfg.AmbientCtx,
		s.cfg.Config,
		s.ClusterSettings(),
		s.sqlExecutor,
		&s.internalMemMetrics,
		&rootSQLMemoryMonitor,
		s.cfg.HistogramWindowInterval(),
		&execCfg,
	)
	s.registry.AddMetricStruct(s.pgServer.Metrics())
	if !s.cfg.UseLegacyConnHandling {
		s.registry.AddMetricStruct(s.pgServer.StatementCounters())
		s.registry.AddMetricStruct(s.pgServer.EngineMetrics())
	}

	sqlExecutor.ExecCfg = &execCfg
	s.execCfg = &execCfg

	s.leaseMgr.SetExecCfg(&execCfg)
	s.leaseMgr.RefreshLeases(s.stopper, s.db, s.gossip)

	s.node.InitLogger(&execCfg)

	return s, nil
}

```
