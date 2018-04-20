

# warning log 

```
W180418 09:53:28.556362 457 storage/engine/rocksdb.go:1755  batch [20/2166/0] commit took 915.153317ms (>500ms):
goroutine 457 [running]:
runtime/debug.Stack(0x2635837c, 0xed2690c97, 0x0)
        /usr/local/go/src/runtime/debug/stack.go:24 +0xa7
github.com/cockroachdb/cockroach/pkg/storage/engine.(*rocksDBBatch).commitInternal(0xc4526aafc0, 0xc4381a9400, 0x24c, 0x400)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/engine/rocksdb.go:1756 +0x128
github.com/cockroachdb/cockroach/pkg/storage/engine.(*rocksDBBatch).Commit(0xc4526aafc0, 0xed2690c00, 0x0, 0x0)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/engine/rocksdb.go:1679 +0x6fe
github.com/cockroachdb/cockroach/pkg/storage.(*Replica).applyRaftCommand(0xc425276000, 0x2612180, 0xc4252711a0, 0xc446cb6938, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/replica.go:5037 +0x4d5
github.com/cockroachdb/cockroach/pkg/storage.(*Replica).processRaftCommand(0xc425276000, 0x2612180, 0xc4252711a0, 0xc446cb6938, 0x8, 0x36, 0x3e659, 0x200000002, 0x3, 0x108, ...)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/replica.go:4742 +0x55a
github.com/cockroachdb/cockroach/pkg/storage.(*Replica).handleRaftReadyRaftMuLocked(0xc425276000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, ...)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/replica.go:3658 +0x12c8 
github.com/cockroachdb/cockroach/pkg/storage.(*Store).processRequestQueue.func1(0x2612180, 0xc4532d5f80, 0xc425276000, 0x2612180)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/store.go:3799 +0x109
github.com/cockroachdb/cockroach/pkg/storage.(*Store).withReplicaForRequest(0xc4204a3800, 0x2612180, 0xc4532d5f80, 0xc462ced520, 0xc464b83ed0, 0x0)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/store.go:3121 +0x135
github.com/cockroachdb/cockroach/pkg/storage.(*Store).processRequestQueue(0xc4204a3800, 0x2612180, 0xc4296c6540, 0x699)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/store.go:3787 +0x229
github.com/cockroachdb/cockroach/pkg/storage.(*raftScheduler).worker(0xc420b66000, 0x2612180, 0xc4296c6540)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/scheduler.go:226 +0x21b
github.com/cockroachdb/cockroach/pkg/storage.(*raftScheduler).Start.func2(0x2612180, 0xc4296c6540)
        /go/src/github.com/cockroachdb/cockroach/pkg/storage/scheduler.go:166 +0x3e
github.com/cockroachdb/cockroach/pkg/util/stop.(*Stopper).RunWorker.func1(0xc425536ab0, 0xc420885b90, 0xc425536a60)
        /go/src/github.com/cockroachdb/cockroach/pkg/util/stop/stopper.go:192 +0xe9
created by github.com/cockroachdb/cockroach/pkg/util/stop.(*Stopper).RunWorker
        /go/src/github.com/cockroachdb/cockroach/pkg/util/stop/stopper.go:185 +0xad
```

 Process requests last. This avoids a scenario where a tick and a
 "quiesce" message are processed in the same iteration and intervening
 raft ready processing unquiesced the replica. Note that request
 processing could also occur first, it just shouldn't occur in between
 ticking and ready processing. It is possible for a tick to be enqueued
 concurrently with the quiescing in which case the replica will
 unquiesce when the tick is processed, but we'll wake the leader in
 that case.

 最后一次处理请求。这样可以避免出现勾号和
 “quiesce”消息在相同的迭代和干预中被处理
 raft准备处理unquiesced复制品。请注意,可以先处理请求
 ，它不应该发生在
 滴答作响和准备好处理之间。tick可能被排队
 同时在这种情况下复制品将静止
 当滴答被处理时不会令人震惊，但我们会把领导叫醒
 那种情况。


https://segmentfault.com/a/1190000008006649
http://www.opscoder.info/ectd-raft-library.html
http://www.opscoder.info/ectd-raft-example.html
