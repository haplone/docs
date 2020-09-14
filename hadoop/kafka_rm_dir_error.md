# kafka_rm_dir_error

rm kafka data dir by accident

reboot kafka, then ok

```
ERROR kafka.server.LogDirFailureChannel: Error while writing to checkpoint file /kafka/data/recovery-point-offset-checkpoint
java.io.FileNotFoundException: /kafka/data/recovery-point-offset-checkpoint.tmp (No such file or directory)
        at java.io.FileOutputStream.open0(Native Method)
        at java.io.FileOutputStream.open(FileOutputStream.java:270)
        at java.io.FileOutputStream.<init>(FileOutputStream.java:213)
        at java.io.FileOutputStream.<init>(FileOutputStream.java:162)
        at kafka.server.checkpoints.CheckpointFile.liftedTree1$1(CheckpointFile.scala:52)
        at kafka.server.checkpoints.CheckpointFile.write(CheckpointFile.scala:50)
        at kafka.server.checkpoints.OffsetCheckpointFile.write(OffsetCheckpointFile.scala:59)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsetsInDir$1$$anonfun$apply$33.apply(LogManager.scala:610)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsetsInDir$1$$anonfun$apply$33.apply(LogManager.scala:608)
        at scala.Option.foreach(Option.scala:257)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsetsInDir$1.apply(LogManager.scala:608)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsetsInDir$1.apply(LogManager.scala:607)
        at scala.Option.foreach(Option.scala:257)
        at kafka.log.LogManager.checkpointLogRecoveryOffsetsInDir(LogManager.scala:607)
        at kafka.log.LogManager.checkpointRecoveryOffsetsAndCleanSnapshot(LogManager.scala:596)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsets$1$$anonfun$apply$32.apply(LogManager.scala:574)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsets$1$$anonfun$apply$32.apply(LogManager.scala:573)
        at scala.Option.foreach(Option.scala:257)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsets$1.apply(LogManager.scala:573)
        at kafka.log.LogManager$$anonfun$checkpointLogRecoveryOffsets$1.apply(LogManager.scala:572)
        at scala.collection.immutable.Map$Map1.foreach(Map.scala:116)
        at kafka.log.LogManager.checkpointLogRecoveryOffsets(LogManager.scala:572)
        at kafka.log.LogManager$$anonfun$startup$3.apply$mcV$sp(LogManager.scala:406)
        at kafka.utils.KafkaScheduler$$anonfun$1.apply$mcV$sp(KafkaScheduler.scala:114)
        at kafka.utils.CoreUtils$$anon$1.run(CoreUtils.scala:63)
        at java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:511)
        at java.util.concurrent.FutureTask.runAndReset(FutureTask.java:308)
        at java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.access$301(ScheduledThreadPoolExecutor.java:180)
        at java.util.concurrent.ScheduledThreadPoolExecutor$ScheduledFutureTask.run(ScheduledThreadPoolExecutor.java:294)
        at java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1149)
        at java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:624)
        at java.lang.Thread.run(Thread.java:748)

INFO kafka.server.ReplicaManager: [ReplicaManager broker=62] Stopping serving replicas in dir /kafka/data

INFO kafka.server.ReplicaFetcherManager: [ReplicaFetcherManager on broker 62] Removed fetcher for partitions Set( topic_names )

WARN kafka.server.ReplicaManager: [ReplicaManager broker=62] While recording the replica LEO, the partition topic_name hasn't been created.

 WARN kafka.server.ReplicaManager: [ReplicaManager broker=62] While recording the replica LEO, the partition topic_name-2 hasn't been created.

INFO kafka.log.LogManager: Stopping serving logs in dir /data2/kafka/data

ERROR kafka.log.LogManager: Shutdown broker because all log dirs in /data2/kafka/data have failed
```

