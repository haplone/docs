# cockroachdb 主要技术点解读

## 核心概念

* gossip 节点发现
* gossip 状态收集：cpu、内存、磁盘等，可用于range split等
* raft 数据同步
* open tracing 分布式调用追踪协议。多语言实现
* lsm算法 将磁盘的随机读，改为顺序读写，提升性能，来源于google 的bigtable
* wal write ahead log 防数据丢失
* SSTable(sorted string table)作为一个连续的kv构成的块
* grpc



## 参考资料
[open tracing协议](https://www.gitbook.com/book/wu-sheng/opentracing-io/details)

[lsm 知乎](https://www.zhihu.com/question/19887265?sort=created)

[raft论文](https://github.com/maemual/raft-zh_cn/blob/master/raft-zh_cn.md)
