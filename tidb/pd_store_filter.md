
## interface 

```go

// Filter is an interface to filter source and target store.
type Filter interface {
	// Return true if the store should not be used as a source store.
	FilterSource(opt Options, store *core.StoreInfo) bool
	// Return true if the store should not be used as a target store.
	FilterTarget(opt Options, store *core.StoreInfo) bool
}

// FilterSource checks if store can pass all Filters as source store.
func FilterSource(opt Options, store *core.StoreInfo, filters []Filter) bool {
	for _, filter := range filters {
		if filter.FilterSource(opt, store) {
			log.Debugf("[filter %T] filters store %v from source", filter, store)
			return true
		}
	}
	return false
}

// FilterTarget checks if store can pass all Filters as target store.
func FilterTarget(opt Options, store *core.StoreInfo, filters []Filter) bool {
	for _, filter := range filters {
		if filter.FilterTarget(opt, store) {
			log.Debugf("[filter %T] filters store %v from target", filter, store)
			return true
		}
	}
	return false
}
```

## implements

* excludedFilter 通过配置map的方式排除特定的store
* blockFilter 将处于block状态的store从balance列表中排除
* stateFilter tombstone状态的store不能作为source，只有up状态的store能作为target
* healthFilter store不能busy，也不能down太久
* disconnectFilter 排除disconnect状态的store
* pendingPeerCountFilter 排除太多pending 状态peer 的store
* snapshotCountFilter 排除太多snapshot的store，不论是sending，receiving，applying状态
* cacheFilter 排除ttl cache中的store
* storageThresholdFilter 排除存储剩余量低于配置的store
* distinctScoreFilter ??
* namespaceFilter 根据rocksdb的namespace进行过滤；待确认
* rejectLeaderFilter 待确认
