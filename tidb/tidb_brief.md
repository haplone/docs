## why tidb

### tp

关于数据库发展的历史，如图一所示在早期，大家主要是使用单机数据库，如Mysql等。这些数据库的性能完全可以满足当时业务的需求。但是自2005年开始，也就是互联网浪潮到来的时候，这些早期的单机数据库就慢慢开始力不从心了。当时Google发表了几篇论文，谈论了其内部使用的Bigtable Mapreduce。然后就有了Redis、HBase 等非关系型数据库，这些数据库实际上已经足以满足当时业务的需求。

![newsql history](newsql_history_brief.jpeg)

直到最近的五六年，我们发现尽管各种NoSQL产品大行其道，但是Mysql依然是不可或缺的。即使是Google，在某一些不能丢数据的场景中，对一些数据的处理依然需要用到ACID，需要用到跨行事务。因此Google在12年的时候发表了一篇名为《Google Spanner》的论文(注：Spanner是在Bigtable之上，用2PC实现了分布式事务)。然后基于Spanner，Google的团队做了F1(注：F1实际上是一个SQL层，支持SQL的语法)，F1用了一些中间状态来屏蔽了单机Mysql可能碰到的阻塞的场景。

### ap

spark sql + thrift server ==> kudu ==> cockroachdb ==> tidb

* spark sql + thrift server 跑批强悍，但频繁查询的场景满足不了
* kudu 底层也是raft + lsm ，遇到的问题是region数量，sql支持依赖impala
* cockroach 啥都好，还json支持非常好；but ... join性能差，使用PostgreSQL语法

#### kudu
![kudu position](../kudu/kudu_hdfs_hbase.jpg)


![kudu architecture](../kudu/kudu_architecture.png)

![how kudu write data](../kudu/kudu_write_data.jpeg)

#### cockroach

![cockroach db](../cockroachdb/media/architecture.png)

## tidb architecture

![tidb all architecture](tidb_all_architecture.jpeg)

![how we use](tidb_all_architecture2.png)

我们日常说的tidb其实是包含3大组件： tidb、pd、tikv。

* `tidb`: SQL layer，使用的mysql语法;负责sql解析、计算；golang实现
* `pd`: 负责kv的调度，统一授时；golang实现
* `tikv`: 基于raft实现的统一kv存储；数据最终存储的地方；rust实现

### tikv

![tikv overview](tikv_overview.jpeg)

换个大家熟悉的角度看

![tikv overview2](tikv_architure.jpg)

Tikv简单说，就是使用raft在单机rocksdb的基础上，做一个分布式的kv存储。以region为单位对数据分组后，通过raft保证数据的外部一致性。

Tikv 跟pd的关系： Tikv主动上报tikv实例信息、每个raft组的信息给pd，pd只是接收。

TiKV 也是一个非常复杂的系统，这块我们会重点介绍，主要包括：

* Raftstore，主要是Raft的实现，包括 Multi-Raft。是etcd版raft使用rust实现的一个版本。
* Storage，Multiversion concurrency control (MVCC)，基于 Percolator 的分布式事务的实现，数据在 engine 里面的存储方式，engine 操作相关的 API 等。
* Server，TiKV 的 gRPC API，以及不同函数执行流程。
* Coprocessor，TiKV 处理 TiDB 的下推请求，如何通过不同的表达式进行数据读取以及计算的。
* PD， TiKV 跟 PD 进行交互。
* Import， TiKV 处理大量数据的导入，以及跟 TiDB 数据导入工具 lightning 交互的。
* Util， TiKV 使用的基本功能库。

### pd

PD主要负责元数据的管理。PD的主要作用：

* 管理TiKV(一段范围key的数据)，通过PD,可以获取每一段数据具体的位置；
* F1中提到的授时，即用于保证分布式数据库的外部一致性。

![pd archtecture](pd_archtecture.jpeg)

### tidb

![tidb architecture](tidb_architure.jpg)

![tidb sql layer](tidb_sql_layer.jpg)

SQL本身实际上就是一串字符串，所以我们首先会有一个解析器。语法解析器(Parser)的实现本身不难，但是却极其繁琐。这主要是因为Mysql比较奇葩的语法，Mysql在SQL标准的基础上增加了很多自己的语法，因此pingcap这一模块的实现耗时极长，有一两个月之久。

但由于Parser只做语法规则的检查，SQL语句本身的逻辑却没法保证，因此在Parser后紧跟了Validate模块来检查SQL语句的逻辑意义。

在Validate模块后是一个类型推断。之所以有这个模块，主要是因为同一个SQL语句在不同的上下文环境有不同的语义，因此我们不得不根据上下文来进行类型推断。

接下来是优化器，这个模块pingcap有一个专门的团队来负责。优化器主要分为逻辑优化器和物理优化器。逻辑优化器主要是基于关系代数，比如子查询去关联等。然后是和底层的物理存储层打交道的物理优化器，索引等就是在这一块实现的。

后面还有一个分布式执行器，它主要是利用MPP模型来实现的，这也是分布式数据库比单机数据库要快的主要原因。

## 引用

[Go 在 TiDB 的实践](http://www.sohu.com/a/220085058_657921)

[kudu overview](https://kudu.apache.org/overview.html)

[Kudu:支持快速分析的新型Hadoop存储系统](https://bigdata.163.com/product/article/1)

[TiKV 源码解析系列文章（一）序](https://pingcap.com/blog-cn/tikv-source-code-reading-1/)