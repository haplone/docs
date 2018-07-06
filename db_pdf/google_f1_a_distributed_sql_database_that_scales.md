

# abstract

F1 is a distributed relational database system built at
Google to support the AdWords business. F1 is a hybrid
database that combines high availability, the scalability of
NoSQL systems like Bigtable, and the consistency and usability of traditional SQL databases. F1 is built on Spanner, which provides synchronous cross-datacenter replication and strong consistency. Synchronous replication implies higher commit latency, but we mitigate that latency
by using a hierarchical schema model with structured data
types and through smart application design. F1 also includes a fully functional distributed SQL query engine and
automatic change tracking and publishing.


F1是一个构建于的分布式关系数据库系统Google支持AdWords业务。
F1是混合数据库，实现了高可用性，像Bigtable这样的NoSQL系统的可扩展性，以及传统SQL数据库的一致性和可用性。 F1基于Spanner构建，可提供同步跨数据中心复制和强一致性。 同步复制意味着更高的提交延迟，但我们会减少延迟通过使用具有结构化数据的分层模式模型
类型和通过智能应用程序设计。 F1还包括一个功能齐全的分布式SQL查询引擎和
自动更改跟踪和发布。

# introduction

F1 is a fault-tolerant globally-distributed OLTP and
OLAP database built at Google as the new storage system
for Google’s AdWords system. It was designed to replace a
sharded MySQL implementation that was not able to meet
our growing scalability and reliability

F1是一个容错的全局分布式OLTP和
OLAP数据库是在Google上构建的新存储系统
适用于Google的AdWords系统。 它旨在取代一个
分片的MySQL实现无法满足
我们不断增长的可扩展性和可靠


The key goals of F1’s design are:

* **Scalability**: The system must be able to scale up,
trivially and transparently, just by adding resources.
Our sharded database based on MySQL was hard to
scale up, and even more difficult to rebalance. Our
users needed complex queries and joins, which meant
they had to carefully shard their data, and resharding
data without breaking applications was challenging.


* **可扩展性**：系统必须能够扩展，通过添加资源，简单而透明地进行。我们基于MySQL的分片数据库很难扩大规模，甚至更难以重新平衡。 我们的用户需要复杂的查询和连接，这意味着他们不得不仔细地对他们的数据进行分片并重新分析数据不破坏应用程序是一项挑战

* **Availability**: The system must never go down for any
reason – datacenter outages, planned maintenance,
schema changes, etc. The system stores data for
Google’s core business. Any downtime has a significant revenue impact.

* **高可用**：系统绝对不能用于任何系统原因 - 数据中心中断，计划维护，架构更改等。系统存储数据谷歌的核心业务。 任何停机都会对收入产生重大影响。

* **Consistency**: The system must provide ACID transactions, and must always present applications with consistent and correct data.
Designing applications to cope with concurrency anomalies in their data is very error-prone, time-consuming, and ultimately not worth the performance gains.

* **一致性**：系统必须提供ACID事务，并且必须始终为应用程序提供一致且正确的数据。
设计应用程序以处理其数据中的并发异常非常容易出错，耗时，并且最终不值得获得性能提升。


* **Usability**: The system must provide full SQL query
support and other functionality users expect from a
SQL database. Features like indexes and ad hoc query
are not just nice to have, but absolute requirements
for our business.

* **可用性**：系统必须提供完整的SQL查询用户期望的支持和其他功能SQL数据库。 索引和即席查询等功能不仅仅是好的，而是绝对的要求为了我们的业务。


Recent publications have suggested that these design goals
are mutually exclusive [5, 11, 23]. A key contribution of this
paper is to show how we achieved all of these goals in F1’s
design, and where we made trade-offs and sacrifices. The
name F1 comes from genetics, where a Filial 1 hybrid is the
first generation offspring resulting from a cross mating of
distinctly different parental types. The F1 database system
is indeed such a hybrid, combining the best aspects of traditional relational databases and scalable NoSQL systems like Bigtable [6].

最近的出版物提出了这些设计目标
相互排斥[5,11,23]。 这是一个关键的贡献
论文将展示我们如何在F1中实现所有这些目标
设计，以及我们做出权衡和牺牲的地方。该
名称F1来自遗传学，其中Filial 1杂交是
第一代后代由交叉交配产生
明显不同的父母类型。 F1数据库系统
确实是这样一种混合体，结合了传统关系数据库的最佳方面和可扩展的NoSQL系统，如Bigtable [6]。


F1 is built on top of Spanner [7], which provides extremely
scalable data storage, synchronous replication, and strong
consistency and ordering properties. F1 inherits those features from Spanner and adds several more:

F1建立在Spanner [7]的基础之上，非常提供可扩展的数据存储，同步复制和强大一致性和排序属性。 F1从Spanner继承了这些功能，并添加了几个：

* Distributed SQL queries, including joining data from external data sources
* Transactionally consistent secondary indexes
* Asynchronous schema changes including database reorganizations
* Optimistic transactions
* Automatic change history recording and publishing

* 分布式SQL查询，包括连接外部数据源的数据
* 事务上一致的二级索引
* 异步模式更改，包括数据库重组
* 乐观的事务
* 自动更改历史记录和发布

Our design choices in F1 result in higher latency for typical reads and writes. We have developed techniques to hide that increased latency, and we found that user-facing transactions can be made to perform as well as in our previous MySQL system:

我们在F1中的设计选择典型读写，而更高延迟。 我们已经开发了隐藏增加的延迟的技术，并且我们发现面向用户的事务可以像以前的MySQL系统一样执行：

* An F1 schema makes data clustering explicit, using tables with hierarchical relationships and columns with structured data types. This clustering improves data locality and reduces the number and cost of RPCs required to read remote data.

* 使用具有层次关系的表和具有结构化数据类型的列，F1模式使数据集群显式化。 此群集可以改善数据的位置，并减少读取远程数据所需的RPC的数量和成本。

*  F1 users make heavy use of batching, parallelism and asynchronous reads. We use a new ORM (object-relational mapping) library that makes these concepts explicit. This places an upper bound on the number of RPCs required for typical application-level operations, making those operations scale well by default.

* F1用户大量使用批处理，并行和异步读取。 我们使用一个新的ORM（对象关系映射）库来使这些概念显式化。 这为典型应用程序级操作所需的RPC数量设置了上限，默认情况下这些操作可以很好地扩展。

The F1 system has been managing all AdWords advertising campaign data in production since early 2012. AdWords is a vast and diverse ecosystem including 100s of applications and 1000s of users, all sharing the same database. This database is over 100 TB, serves up to hundreds of thousands of requests per second, and runs SQL queries that scan tens of trillions of data rows per day. Availability reaches five nines, even in the presence of unplanned outages, and observable latency on our web applications has not increased compared to the old MySQL system.


自2012年初以来，F1系统一直在管理所有AdWords广告活动数据.AdWords是一个庞大而多样的生态系统，包括100个应用程序和1000个用户，所有这些都共享同一个数据库。 该数据库超过100 TB，每秒可处理数十万个请求，并运行每天扫描数十万亿个数据行的SQL查询。 即使存在计划外中断，可用性也达到了五个九，与旧的MySQL系统相比，我们的Web应用程序的可观察延迟没有增加。

We discuss the AdWords F1 database throughout this paper as it was the original and motivating user for F1. Several other groups at Google are now beginning to deploy F1.

我们在本文中讨论了AdWords F1数据库，因为它是F1的原始用户和激励用户。 谷歌的其他几个团体现在开始部署F1。

# basic architecture

Users interact with F1 through the F1 client library. Other tools like the command-line ad-hoc SQL shell are implemented using the same client. The client sends requests to one of many F1 servers, which are responsible for reading and writing data from remote data sources and coordinating query execution. Figure 1 depicts the basic architecture and the communication between components.

用户通过F1客户端库与F1交互。 其他工具（如命令行ad-hoc SQL shell）使用同一客户端实现。 客户端向许多F1服务器之一发送请求，这些服务器负责从远程数据源读取和写入数据并协调查询执行。 图1描述了基本架构和组件之间的通信。


Because of F1’s distributed architecture, special care must be taken to avoid unnecessarily increasing request latency. For example, the F1 client and load balancer prefer to connect to an F1 server in a nearby datacenter whenever possible. However, requests may transparently go to F1 servers in remote datacenters in cases of high load or failures.

由于F1的分布式架构，必须特别注意避免不必要地增加请求延迟。 例如，F1客户端和负载均衡器更愿意尽可能连接到附近数据中心的F1服务器。 但是，在高负载或故障的情况下，请求可以透明地转到远程数据中心的F1服务器。


F1 servers are typically co-located in the same set of datacenters as the Spanner servers storing the data. This colocation ensures that F1 servers generally have fast access to the underlying data. For availability and load balancing,F1 servers can communicate with Spanner servers outside their own datacenter when necessary. The Spanner servers in each datacenter in turn retrieve their data from the Colossus File System (CFS) [14] in the same datacenter. Unlike Spanner, CFS is not a globally replicated service and therefore Spanner servers will never communicate with remote CFS instances.

F1服务器通常与存储数据的Spanner服务器位于同一组数据中心中。 此共置确保F1服务器通常可以快速访问基础数据。 对于可用性和负载平衡，F1服务器可以必要时与其自己的数据中心外的Spanner服务器通信。 每个数据中心的Spanner服务器依次从同一数据中心的Colossus文件系统（CFS）[14]中检索数据。 与Spanner不同，CFS不是全局复制服务，因此Spanner服务器永远不会与远程CFS实例通信。


F1 servers are mostly stateless, allowing a client to communicate with a different F1 server for each request. The one exception is when a client uses pessimistic transactions and must hold locks. The client is then bound to one F1 server for the duration of that transaction. F1 transactions are described in more detail in Section 5. F1 servers can be quickly added (or removed) from our system in response to the total load because F1 servers do not own any data and hence a server addition (or removal) requires no data movement.

F1服务器大多数是无状态的，允许客户端为每个请求与不同的F1服务器进行通信。 一个例外是客户端使用悲观事务并且必须持有锁。 然后，客户端在该事务期间绑定到一个F1服务器。 F1事务在第5节中有更详细的描述.F1服务器可以从我们的系统中快速添加（或删除）以响应总负载，因为F1服务器不拥有任何数据，因此服务器添加（或删除）不需要数据 运动。


An F1 cluster has several additional components that allow for the execution of distributed SQL queries. Distributed execution is chosen over centralized execution when the query planner estimates that increased parallelism will reduce query processing latency. The shared slave pool consists of F1 processes that exist only to execute parts of distributed query plans on behalf of regular F1 servers. Slave pool membership is maintained by the F1 master, which monitors slave process health and distributes the list of available slaves to F1 servers. F1 also supports large-scale data processing through Google’s MapReduce framework [10]. For performance reasons, MapReduce workers are allowed to communicate directly with Spanner servers to extract data in bulk (not shown in the figure). Other clients perform reads and writes exclusively through F1 servers.

F1集群有几个附加组件，允许执行分布式SQL查询。 当查询计划程序估计增加的并行性将减少查询处理延迟时，选择分布式执行而不是集中执行。 共享从属池由F1进程组成，这些进程仅代表常规F1服务器执行分布式查询计划的一部分。 从属池成员资格由F1主服务器维护，它监视从属进程运行状况并将可用从属服务器列表分发给F1服务器。 F1还支持通过Google的MapReduce框架进行大规模数据处理[10]。 出于性能原因，允许MapReduce工作者直接与Spanner服务器通信以批量提取数据（图中未显示）。 其他客户端仅通过F1服务器执行读写操作。

The throughput of the entire system can be scaled up by adding more Spanner servers, F1 servers, or F1 slaves. Since F1 servers do not store data, adding new servers does not involve any data re-distribution costs. Adding new Spanner servers results in data re-distribution. This process is completely transparent to F1 servers (and therefore F1 clients).

通过添加更多Spanner服务器，F1服务器或F1 slaves，可以扩大整个系统的吞吐量。 由于F1服务器不存储数据，因此添加新服务器不会涉及任何数据re-distributed成本。 添加新的Spanner服务器会导致数据重新分配。 此过程对F1服务器（以及F1客户端）完全透明。


The Spanner-based remote storage model and our geographically distributed deployment leads to latency characteristics that are very different from those of regular databases. Because the data is synchronously replicated across multiple datacenters, and because we’ve chosen widely distributed datacenters, the commit latencies are relatively high (50-150 ms). This high latency necessitates changes to the patterns that clients use when interacting with the database. We describe these changes in Section 7.1, and we provide further detail on our deployment choices, and the resulting availability and latency, in Sections 9 and 10.

基于Spanner的远程存储模型和我们的地理分布式部署导致了与常规数据库非常不同的延迟特性。 因为数据是跨多个数据中心同步复制的，并且因为我们选择了广泛分布的数据中心，所以提交延迟相对较高（50-150毫秒）。 这种高延迟需要更改客户端在与数据库交互时使用的模式。 我们在7.1节中描述了这些变化，并在第9节和第10节中提供了有关我们的部署选择以及由此产生的可用性和延迟的更多详细信息。

## Spanner

F1 is built on top of Spanner. Both systems were developed at the same time and in close collaboration. Spanner handles lower-level storage issues like persistence, caching, replication, fault tolerance, data sharding and movement, location lookups, and transactions.

F1建立在spanner之上。 两个系统都是在同一时间和密切合作的基础上开发的。 Spanner处理较低级别的存储问题，如持久性，缓存，复制，容错，数据分片和移动，位置查找和事务。

In Spanner, data rows are partitioned into clusters called directories using ancestry relationships in the schema. Each directory has at least one fragment, and large directories can have multiple fragments. Groups store a collection of directory fragments. Each group typically has one replica tablet per datacenter. Data is replicated synchronously using the Paxos algorithm [18], and all tablets for a group store the same data. One replica tablet is elected as the Paxos leader for the group, and that leader is the entry point for all transactional activity for the group. Groups may also include readonly replicas, which do not vote in the Paxos algorithm and cannot become the group leader.

在Spanner中，数据行通过schema中的祖先关系分区到集群的目录。 每个目录至少有一个片段，大型目录可以有多个片段。 组存储目录片段的集合。 每个组通常在每个数据中心有一个replica tablet。 使用Paxos算法[18]同步复制数据，并且组的所有tablets都存储相同的数据。 一个replica tablet被选为该小组的Paxos领导者，该领导者是该小组所有交易活动的切入点。 组还可以包括只读副本，它们不在Paxos算法中投票而不能成为组长。
