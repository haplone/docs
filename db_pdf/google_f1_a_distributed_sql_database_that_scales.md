F1: A Distributed SQL Database That Scales

# abstract

F1 is a distributed relational database system built at Google to support the AdWords business. F1 is a hybrid database that combines high availability, the scalability of NoSQL systems like Bigtable, and the consistency and usability of traditional SQL databases. F1 is built on Spanner, which provides synchronous cross-datacenter replication and strong consistency. Synchronous replication implies higher commit latency, but we mitigate that latency by using a hierarchical schema model with structured data types and through smart application design. F1 also includes a fully functional distributed SQL query engine and automatic change tracking and publishing.


F1是一个构建于的分布式关系数据库系统Google支持AdWords业务。
F1是混合数据库，实现了高可用性，像Bigtable这样的NoSQL系统的可扩展性，以及传统SQL数据库的一致性和可用性。 F1基于Spanner构建，可提供同步跨数据中心复制和强一致性。 同步复制意味着更高的提交延迟，但我们会减少延迟通过使用具有结构化数据的分层模式模型
类型和通过智能应用程序设计。 F1还包括一个功能齐全的分布式SQL查询引擎和
自动更改跟踪和发布。

# introduction

F11 is a fault-tolerant globally-distributed OLTP and OLAP database built at Google as the new storage system for Google’s AdWords system. It was designed to replace a sharded MySQL implementation that was not able to meet our growing scalability and reliability requirements.

F1是一个容错的全局分布式OLTP和
OLAP数据库是在Google上构建的新存储系统
适用于Google的AdWords系统。 它旨在取代一个
分片的MySQL实现无法满足
我们不断增长的可扩展性和可靠


The key goals of F1’s design are:

* **Scalability**: The system must be able to scale up, trivially and transparently, just by adding resources. Our sharded database based on MySQL was hard to scale up, and even more difficult to rebalance. Our users needed complex queries and joins, which meant they had to carefully shard their data, and resharding data without breaking applications was challenging.


* **可扩展性**：系统必须能够扩展，通过添加资源，简单而透明地进行。我们基于MySQL的分片数据库很难扩大规模，甚至更难以重新平衡。 我们的用户需要复杂的查询和连接，这意味着他们不得不仔细地对他们的数据进行分片并重新分析数据不破坏应用程序是一项挑战

* **Availability**: The system must never go down for any reason – datacenter outages, planned maintenance, schema changes, etc. The system stores data for Google’s core business. Any downtime has a significant revenue impact.

* **高可用**：系统绝对不能用于任何系统原因 - 数据中心中断，计划维护，架构更改等。系统存储数据谷歌的核心业务。 任何停机都会对收入产生重大影响。

* **Consistency**: The system must provide ACID transactions, and must always present applications with consistent and correct data.
Designing applications to cope with concurrency anomalies in their data is very error-prone, time-consuming, and ultimately not worth the performance gains.


* **一致性**：系统必须提供ACID事务，并且必须始终为应用程序提供一致且正确的数据。
设计应用程序以处理其数据中的并发异常非常容易出错，耗时，并且最终不值得获得性能提升。


* **Usability**: The system must provide full SQL query support and other functionality users expect from a SQL database. Features like indexes and ad hoc query are not just nice to have, but absolute requirements for our business.

* **易用性**：系统必须提供完整的SQL查询用户期望的支持和其他功能SQL数据库。 索引和即席查询等功能不仅仅是好的，而是绝对的要求为了我们的业务。


Recent publications have suggested that these design goals are mutually exclusive [5, 11, 23]. A key contribution of this paper is to show how we achieved all of these goals in F1’s design, and where we made trade-offs and sacrifices. The name F1 comes from genetics, where a Filial 1 hybrid is the first generation offspring resulting from a cross mating of distinctly different parental types. The F1 database system is indeed such a hybrid, combining the best aspects of traditional relational databases and scalable NoSQL systems like Bigtable [6].


最近的出版物提出了这些设计目标
相互排斥[5,11,23]。 这是一个关键的贡献
论文将展示我们如何在F1中实现所有这些目标
设计，以及我们做出权衡和牺牲的地方。该
名称F1来自遗传学，其中Filial 1杂交是
第一代后代由交叉交配产生
明显不同的父母类型。 F1数据库系统
确实是这样一种混合体，结合了传统关系数据库的最佳方面和可扩展的NoSQL系统，如Bigtable [6]。


F1 is built on top of Spanner [7], which provides extremely scalable data storage, synchronous replication, and strong consistency and ordering properties. F1 inherits those features from Spanner and adds several more:

F1建立在Spanner [7]的基础之上，非常提供可扩展的数据存储，同步复制和强大一致性和排序属性。 F1从Spanner继承了这些功能，并添加了几个：

* Distributed SQL queries, including joining data from external data sources
* Transactionally consistent secondary indexes
* Asynchronous schema changes including database reorganizations
* Optimistic transactions
* Automatic change history recording and publishing

* 分布式SQL查询，包括join外部数据源的数据
* 事务上一致的二级索引
* 异步模式更改，包括数据库重组
* 乐观的事务
* 更改历史的自动记录和发布

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

Spanner provides serializable pessimistic transactions using strict two-phase locking. A transaction includes multiple reads, taking shared or exclusive locks, followed by a single write that upgrades locks and atomically commits the transaction. All commits are synchronously replicated using Paxos. Transactions are most efficient when updating data co-located in a single group. Spanner also supports transactions across multiple groups, called transaction participants, using a two-phase commit (2PC) protocol on top of Paxos. 2PC adds an extra network round trip so it usually doubles observed commit latency. 2PC scales well up to 10s of participants, but abort frequency and latency increase significantly with 100s of participants [7].

Spanner使用严格的两阶段锁定提供可序列化的悲观事务。 事务包括多次读取，获取共享锁或独占锁，然后是单次写入，用于升级锁并以原子方式提交事务。 使用Paxos同步复制所有提交。 更新共同位于单个组中的数据时，事务处理效率最高。 Spanner还支持跨多个组的事务，称为事务参与者，使用Paxos之上的两阶段提交（2PC）协议。 2PC增加了额外的网络往返，因此它通常会使观察到的提交延迟加倍。 2PC可以很好地扩展到10个参与者，但中止频率和潜伏期显着增加，参与者数量达到100个[7]。

Spanner has very strong consistency and timestamp semantics. Every transaction is assigned a commit timestamp,and these timestamps provide a global total ordering for commits. Spanner uses a novel mechanism to pick globally ordered timestamps in a scalable way using hardware clocks deployed in Google datacenters. Spanner uses these timestamps to provide multi-versioned consistent reads, including snapshot reads of current data, without taking read locks. For guaranteed non-blocking, globally consistent reads, Spanner provides a global safe timestamp, below which no in-flight or future transaction can possibly commit. The global safe timestamp typically lags current time by 5-10 seconds. Reads at this timestamp can normally run on any replica tablet, including readonly replicas, and they never block behind running transactions.

Spanner具有非常强的一致性和时间戳语义。 为每个事务分配一个提交时间戳，这些时间戳为提交提供全局总排序。 Spanner使用一种新颖的机制，使用部署在Google数据中心的硬件时钟以可扩展的方式选择全局有序时间戳。 Spanner使用这些时间戳来提供多版本的一致性读取，包括当前数据的快照读取，而不需要读取锁定。 对于有保证的非阻塞，全局一致的读取，Spanner提供了一个全局安全时间戳，低于该时间戳，没有空中或将来的事务可能提交。 全局安全时间戳通常滞后当前时间5-10秒。 此时间戳的读取通常可以在任何replica tablet 上运行，包括只读副本，并且它们永远不会阻止正在运行的事务。


# data model

## hierachical schema


The F1 data model is very similar to the Spanner data model. In fact, Spanner’s original data model was more like Bigtable, but Spanner later adopted F1’s data model. At the logical level, F1 has a relational schema similar to that of a traditional RDBMS, with some extensions including explicit table hierarchy and columns with Protocol Buffer data types.

F1数据模型与Spanner数据模型非常相似。 事实上，Spanner的原始数据模型更像是Bigtable，但Spanner后来采用了F1的数据模型。 在逻辑层面，F1具有类似于传统RDBMS的关系模式，其中一些扩展包括显式表层次结构和具有Protocol Buffer数据类型的列。

Logically, tables in the F1 schema can be organized into a hierarchy. Physically, F1 stores each child table clustered with and interleaved within the rows from its parent table. Tables from the logical schema cannot be arbitrarily interleaved: the child table must have a foreign key to its parent table as a prefix of its primary key. For example, the AdWords schema contains a table Customer with primary key (CustomerId), which has a child table Campaign with primary key (CustomerId, CampaignId), which in turn has a child table AdGroup with primary key (CustomerId,CampaignId, AdGroupId). A row of the root table in the hierarchy is called a root row. All child table rows corresponding to a root row are clustered together with that root row in a single Spanner directory, meaning that cluster is normally stored on a single Spanner server. Child rows are stored under their parent row ordered by primary key. Figure 2 shows an example.

逻辑上，F1模式中的表可以组织成层次结构。 物理上，F1存储每个子表，这些子表在其父表的行中聚集并交错。 逻辑模式中的表不能任意交错：子表必须具有其父表的外键作为其主键的前缀。 例如，AdWords架构包含一个带有主键的客户表CustomerId（CustomerId），该表具有带主键的子表Campaign（CustomerId，CampaignId），后者又具有带主键的子表AdGroup（CustomerId，CampaignId，AdGroupId）。 层次结构中的根表的一行称为根行。 对应于根行的所有子表行与该根行一起聚集在一个Spanner目录中，这意味着该集群通常存储在单个Spanner服务器上。 子行存储在按主键排序的父行下。 图2显示了一个示例。

The hierarchically clustered physical schema has several advantages over a flat relational schema. Consider the corresponding traditional schema, also depicted in Figure 2. In this traditional schema, fetching all Campaign and AdGroup records corresponding to a given CustomerId would take two sequential steps, because there is no direct way to retrieve AdGroup records by CustomerId. In the F1 version of the schema, the hierarchical primary keys allow the fetches of Campaign and AdGroup records to be started in parallel, because both tables are keyed by CustomerId. The primary key prefix property means that reading all AdGroups for a particular Customer can be expressed as a single range read, rather than reading each row individually using an index. Furthermore, because the tables are both stored in primary key order, rows from the two tables can be joined using a simple ordered merge. Because the data is clustered into a single directory, we can read it all in a single Spanner request. All of these properties of a hierarchical schema help mitigate the latency effects of having remote data.

与平面关系模式相比，分层集群物理模式具有几个优点。考虑相应的传统模式，如图2所示。在此传统模式中，获取与给定CustomerId相对应的所有Campaign和AdGroup记录将需要两个连续步骤，因为没有直接方法可以通过CustomerId检索AdGroup记录。在架构的F1版本中，分层主键允许并行启动Campaign和AdGroup记录的提取，因为两个表都由CustomerId键控。主键前缀属性意味着读取特定客户的所有广告组可以表示为单个范围读取，而不是使用索引单独读取每一行。此外，因为表都以主键顺序存储，所以可以使用简单的有序合并来连接来自两个表的行。因为数据被聚集到一个目录中，所以我们可以在一个Spanner请求中读取它。分层模式的所有这些属性有助于减轻具有远程数据的延迟效应。


Hierarchical clustering is especially useful for updates, since it reduces the number of Spanner groups involved in a transaction. Because each root row and all of its descendant rows are stored in a single Spanner directory, transactions restricted to a single root will usually avoid 2PC and the associated latency penalty, so most applications try to use single-root transactions as much as possible. Even when doing transactions across multiple roots, it is important to limit the number of roots involved because adding more participants generally increases latency and decreases the likelihood of a successful commit.

分层聚类对于更新尤其有用，因为它减少了事务中涉及的Spanner组的数量。 由于每个根行及其所有后代行都存储在单个Spanner目录中，因此限制为单个根的事务通常会避免2PC和相关的延迟惩罚，因此大多数应用程序尝试尽可能多地使用单根事务。 即使在跨多个根进行事务时，限制所涉及的根数也很重要，因为添加更多参与者通常会增加延迟并降低成功提交的可能性。


Hierarchical clustering is not mandatory in F1. An F1 schema often has several root tables, and in fact, a completely flat MySQL-style schema is still possible. Using hierarchy however, to the extent that it matches data semantics, is highly beneficial. In AdWords, most transactions are typically updating data for a single advertiser at a time, so we made the advertiser a root table (Customer) and clustered related tables under it. This clustering was critical to achieving acceptable latency.

在F1中，分层聚类不是必需的。 F1模式通常有几个根表，事实上，仍然可以使用完全平坦的MySQL样式模式。 然而，使用层次结构，在与数据语义匹配的程度上，是非常有益的。 在AdWords中，大多数交易通常一次更新单个广告客户的数据，因此我们将广告客户设为根表（客户）和其下的群集相关表。 这种群集对于实现可接受的延迟至关重要。

## Protocol Buffer

The F1 data model supports table columns that contain structured data types. These structured types use the schema and binary encoding format provided by Google’s open source Protocol Buffer [16] library. Protocol Buffers have typed fields that can be required, optional, or repeated; fields can also be nested Protocol Buffers. At Google, Protocol Buffers are ubiquitous for data storage and interchange between applications. When we still had a MySQL schema, users often had to write tedious and error-prone transformations between database rows and in-memory data structures. Putting protocol buffers in the schema removes this impedance mismatch and gives users a universal data structure they can use both in the database and in application code.

F1数据模型支持包含结构化数据类型的表列。 这些结构化类型使用Google的开源协议缓冲区[16]库提供的模式和二进制编码格式。 Protocol Buffer 具有可以是必需的，可选的或重复的类型字段; 字段也可以是嵌套的Protocol Buffers。 在Google，Protocol Buffer无处不在，用于数据存储和应用程序之间的交换。 当我们仍然拥有MySQL模式时，用户经常不得不在数据库行和内存数据结构之间编写繁琐且容易出错的转换。 将Protocol Buffer放在模式中会消除阻抗不匹配，并为用户提供可在数据库和应用程序代码中使用的通用数据结构。


Protocol Buffers allow the use of repeated fields. In F1 schema designs, we often use repeated fields instead of child tables when the number of child records has a low upper bound. By using repeated fields, we avoid the performance overhead and complexity of storing and joining multiple child records. The entire protocol buffer is effectively treated as one blob by Spanner. Aside from performance impacts, Protocol Buffer columns are more natural and reduce semantic complexity for users, who can now read and write their logical business objects as atomic units, without having to  hink about materializing them using joins across several tables. The use of Protocol Buffers in F1 SQL is described in Section 8.7.


协议缓冲区允许使用重复的字段。 在F1模式设计中，当子记录的数量上限较低时，我们经常使用重复的字段而不是子表。 通过使用重复字段，我们避免了存储和连接多个子记录的性能开销和复杂性。 整个协议缓冲区被Spanner有效地视为一个blob。 除了性能影响之外，协议缓冲区列更自然，并且降低了用户的语义复杂性，用户现在可以将逻辑业务对象作为原子单元进行读写，而不必使用跨多个表的连接来实现它们。 第8.7节描述了在F1 SQL中使用Protocol Buffers。

Many tables in an F1 schema consist of just a single Protocol Buffer column. Other tables split their data across a handful of columns, partitioning the fields according to access patterns. Tables can be partitioned into columns to group together fields that are usually accessed together, to separate fields with static and frequently updated data, to allow specifying different read/write permissions per column, or to allow concurrent updates to different columns. Using fewer columns generally improves performance in Spanner where there can be high per-column overhead.


F1模式中的许多表只包含一个Protocol Buffer列。 其他表将数据分成几列，根据访问模式对字段进行分区。 可以将表分区为列，将通常一起访问的字段组合在一起，将字段与静态和频繁更新的数据分开，以允许为每列指定不同的读/写权限，或允许对不同列进行并发更新。 使用较少的列通常可以提高Spanner的性能，因为每个列的开销很高。


## indexing

All indexes in F1 are transactional and fully consistent. Indexes are stored as separate tables in Spanner, keyed by a concatenation of the index key and the indexed table’s primary key. Index keys can be either scalar columns or fields extracted from Protocol Buffers (including repeated fields). There are two types of physical storage layout for F1 indexes: local and global.

F1中的所有索引都是事务性的并且完全一致。 索引作为单独的表存储在Spanner中，由索引键和索引表的主键串联键入。 索引键可以是标量列，也可以是从协议缓冲区中提取的字段（包括重复字段）。 F1索引有两种类型的物理存储布局：本地和全局。


Local index keys must contain the root row primary key as a prefix. For example, an index on (CustomerId, Keyword) used to store unique keywords for each customer is a local index. Like child tables, local indexes are stored in the same Spanner directory as the root row. Consequently, the index entries of local indexes are stored on the same Spanner server as the rows they index, and local index updates add little additional cost to any transaction.

本地索引键必须包含根行主键作为前缀。 例如，用于为每个客户存储唯一关键字的（CustomerId，Keyword）索引是本地索引。 与子表一样，本地索引与根行存储在同一个Spanner目录中。 因此，本地索引的索引条目与它们索引的行存储在同一个Spanner服务器上，而本地索引更新为任何事务添加了很少的额外成本。

In contrast, global index keys do not include the root row primary key as a prefix and hence cannot be co-located with the rows they index. For example, an index on (Keyword) that maps from all keywords in the database to Customers that use them must be global. Global indexes are often large and can have high aggregate update rates. Consequently, they are sharded across many directories and stored on multiple Spanner servers. Writing a single row that updates a global index requires adding a single extra participant to a transaction, which means the transaction must use 2PC, but that is a reasonable cost to pay for consistent global indexes.

相反，全局索引键不包括根行主键作为前缀，因此不能与它们索引的行共同定位。 例如，从数据库中的所有关键字映射到使用它们的客户的（关键字）索引必须是全局的。 全局索引通常很大，并且可以具有高聚合更新率。 因此，它们在多个目录中分片并存储在多个Spanner服务器上。 编写更新全局索引的单行需要向事务添加一个额外的参与者，这意味着事务必须使用2PC，但这是支付一致全局索引的合理成本。


Global indexes work reasonably well for single-row up dates, but can cause scaling problems for large transactions. Consider a transaction that inserts 1000 rows. Each row requires adding one or more global index entries, and those index entries could be arbitrarily spread across 100s of index directories, meaning the 2PC transaction would have 100s of participants, making it slower and more error-prone. Therefore, we use global indexes sparingly in the schema, and encourage application writers to use small transactions when bulk inserting into tables with global indexes.

对于单行更新日期，全局索引运行良好，但可能导致大型事务的扩展问题。 考虑插入1000行的事务。 每行需要添加一个或多个全局索引条目，并且这些索引条目可以任意分布在100个索引目录中，这意味着2PC事务将有100个参与者，使其更慢且更容易出错。 因此，我们在模式中谨慎使用全局索引，并鼓励应用程序编写者在批量插入具有全局索引的表时使用小事务。


Megastore [3] makes global indexes scalable by giving up consistency and supporting only asynchronous global indexes. We are currently exploring other mechanisms to make global indexes more scalable without compromising consistency.

Megastore [3]通过放弃一致性并仅支持异步全局索引来使全局索引可扩展。 我们目前正在探索其他机制，以使全局索引更具可扩展性而不会影响一致性。


# schema changes

The AdWords database is shared by thousands of users and is under constant development. Batches of schema changes are queued by developers and applied daily. This database is mission critical for Google and requires very high availability. Downtime or table locking during schema changes (e.g. adding indexes) is not acceptable.

We have designed F1 to make all schema changes fully non-blocking. Several aspects of the F1 system make non-blocking schema changes particularly challenging:

AdWords数据库由数千名用户共享，并且正在不断发展。 批量的架构更改由开发人员排队并每天应用。 该数据库对Google而言至关重要，需要非常高的可用性。 架构更改期间的停机或表锁定（例如添加索引）是不可接受的。

我们设计了F1以使所有架构更改完全无阻塞。 F1系统的几个方面使非阻塞模式更改特别具有挑战性：

* F1 is a massively distributed system, with servers in multiple datacenters in distinct geographic regions.

* Each F1 server has a schema locally in memory. It is not practical to make an update occur atomically across all servers.

* Queries and transactions must continue on all tables, even those undergoing schema changes.

* System availability and latency must not be negatively impacted during schema changes.

* F1是一个大规模分布式系统，服务器位于不同地理区域的多个数据中心。

* 每个F1服务器在内存中都有一个架构。 在所有服务器上以原子方式进行更新是不切实际的。

* 查询和事务必须在所有表上继续，即使是那些正在进行模式更改的表。

* 在架构更改期间，系统可用性和延迟不得受到负面影响。

Because F1 is massively distributed, even if F1 had a global F1 server membership repository, synchronous schema change across all servers would be very disruptive to response times. To make changes atomic, at some point, servers would have to block transactions until confirming all other servers have received the change. To avoid this, F1 schema changes are applied asynchronously, on different F1 servers at different times. This implies that two F1 servers may update the database concurrently using different schemas.

由于F1是大规模分布式的，即使F1拥有全局F1服务器成员资格存储库，所有服务器上的同步模式更改也会对响应时间造成极大的破坏。 要使更改成为原子，在某些时候，服务器必须阻止事务，直到确认所有其他服务器都已收到更改。 为避免这种情况，F1架构更改将在不同时间在不同的F1服务器上异步应用。 这意味着两个F1服务器可以使用不同的模式同时更新数据库。

If two F1 servers update the database using different schemas that are not compatible according to our schema change algorithms, this could lead to anomalies including database corruption. We illustrate the possibility of database corruption using an example. Consider a schema change from schema S 1 to schema S 2 that adds index I on table T . Because the schema change is applied asynchronously on different F1 servers, assume that server M 1 is using schema S 1 and server M 2 is using schema S 2 . First, server M 2 inserts a new row r, which also adds a new index entry I(r) for row r. Subsequently, row r is deleted by server M 1 . Because the server is using schema S 1 and is not aware of index I, the server deletes row r, but fails to delete the index entry I(r). Hence, the database becomes corrupt. For example, an index scan on I would return spurious data corresponding to the deleted row r.


如果两个F1服务器使用根据我们的架构更改算法不兼容的不同架构更新数据库，则可能导致包括数据库损坏在内的异常。 我们使用一个例子来说明数据库损坏的可能性。 考虑从模式S 1到模式S 2的模式更改，它在表T上添加索引I. 由于架构更改是在不同的F1服务器上异步应用的，因此假设服务器M 1使用架构S 1而服务器M 2正在使用架构S 2。 首先，服务器M 2插入新的行r，其还为行r添加新的索引条目I（r）。 随后，服务器M1删除行r。 由于服务器使用模式S1并且不知道索引I，因此服务器删除行r，但无法删除索引条目I（r）。 因此，数据库变得腐败。 例如，对I的索引扫描将返回与删除的行r相对应的伪数据。

We have implemented a schema change algorithm that prevents anomalies similar to the above by

1. Enforcing that across all F1 servers, at most two different schemas are active. Each server uses either the current or next schema. We grant leases on the schema and ensure that no server uses a schema after lease expiry.

1. 在所有F1服务器中执行该操作，最多有两个不同的模式处于活动状态。 每个服务器使用当前或下一个模式。 我们在架构上授予租约，并确保租约到期后没有服务器使用schema。

2. Subdividing each schema change into multiple phases where consecutive pairs of phases are mutually compatible and cannot cause anomalies. In the above example, we first add index I in a mode where it only executes delete operations. This prohibits server M 1 from adding I(r) into the database. Subsequently, we upgrade index I so servers perform all write operations. Then we initiate a MapReduce to backfill index entries for all rows in table T with carefully constructed transactions to handle concurrent writes. Once complete, we make index I visible for normal read perations.

2. 将每个模式更改细分为多个阶段，其中连续的阶段对相互兼容且不会导致异常。 在上面的示例中，我们首先在一个只执行删除操作的模式下添加索引I. 这禁止服务器M 1将I（r）添加到数据库中。 随后，我们升级索引I，以便服务器执行所有写入操作。 然后我们启动MapReduce来回填表T中所有行的索引条目，并使用精心构造的事务来处理并发写入。 完成后，我们将索引I显示为正常读取操作。

The full details of the schema change algorithms are covered in [20].


# transactions

The AdWords product ecosystem requires a data store that supports ACID transactions. We store financial data and have hard requirements on data integrity and consistency. We also have a lot of experience with eventual consistency systems at Google. In all such systems, we find developers spend a significant fraction of their time building extremely complex and error-prone mechanisms to cope with eventual consistency and handle data that may be out of date. We think this is an unacceptable burden to place on developers and that consistency problems should be solved at the database level. Full transactional consistency is one of the most important properties of F1.


Each F1 transaction consists of multiple reads, optionally followed by a single write that commits the transaction. F1 implements three types of transactions, all built on top of Spanner’s transaction support:

1. **Snapshot transactions.** These are read-only transactions with snapshot semantics, reading repeatable data as of a fixed Spanner snapshot timestamp. By default, snapshot transactions read at Spanner’s global safe timestamp, typically 5-10 seconds old, and read from a local Spanner replica. Users can also request a specific timestamp explicitly, or have Spanner pick the current timestamp to see current data. The latter option may have higher latency and require remote RPCs.

Snapshot transactions are the default mode for SQL queries and for MapReduces. Snapshot transactions allow multiple client servers to see consistent views of the entire database at the same timestamp.


2. **Pessimistic transactions.** These transactions map directly on to Spanner transactions [7]. Pessimistic transactions use a stateful communications protocol that requires holding locks, so all requests in a single pessimistic transaction get directed to the same F1 server. If the F1 server restarts, the pessimistic transaction aborts. Reads in pessimistic transactions can request either shared or exclusive locks.

3. **Optimistic transactions.** Optimistic transactions consist of a read phase, which can take arbitrarily long and never takes Spanner locks, and then a short write phase. To detect row-level conflicts, F1 returns with each row its last modification timestamp, which is stored in a hidden lock column in that row. The new commit timestamp is automatically written into the lock column whenever the corresponding data is updated (in either pessimistic or optimistic transactions). The client library collects these timestamps, and passes them back to an F1 server with the write that commits the transaction. The F1 server creates a short-lived Spanner pessimistic transaction and re-reads the last modification timestamps for all read rows. If any of the re-read timestamps differ from what was passed in by the client, there was a conflicting update, and F1 aborts the transaction. Otherwise, F1 sends the writes on to Spanner to finish the commit.

F1 clients use optimistic transactions by default. Optimistic transactions have several benefits:

* Tolerating misbehaved clients. Reads never hold locks and never conflict with writes. This avoids any problems caused by badly behaved clients who run long transactions or abandon transactions without aborting them.

* Long-lasting transactions. Optimistic transactions can be arbitrarily long, which is useful in some cases. For example, some F1 transactions involve waiting for enduser interaction. It is also hard to debug a transaction that always gets aborted while single-stepping. Idle transactions normally get killed within ten seconds to avoid leaking locks, which means long-running pessimistic transactions often cannot commit.

* Server-side retriability. Optimistic transaction commits are self-contained, which makes them easy to retry transparently in the F1 server, hiding most transient Spanner errors from the user. Pessimistic transactions cannot be retried by F1 servers because they require re-running the user’s business logic to reproduce the same locking side-effects.

* Server failover. All state associated with an optimistic transaction is kept on the client. Consequently, the client can send reads and commits to different F1 servers after failures or to balance load.

* Speculative writes. A client may read values outside an optimistic transaction (possibly in a MapReduce), and remember the timestamp used for that read. Then the client can use those values and timestamps in an optimistic transaction to do speculative writes that only succeed if no other writes happened after the original read.
投机

Optimistic transactions do have some drawbacks:

缺点

* Insertion phantoms. Modification timestamps only exist for rows present in the table, so optimistic transactions do not prevent insertion phantoms [13]. Where this is a problem, it is possible to use parent-table locks to avoid phantoms. (See Section 5.1)

* Low throughput under high contention. For example, in a table that maintains a counter which many clients increment concurrently, optimistic transactions lead to many failed commits because the read timestamps are usually stale by write time. In such cases, pessimistic transactions with exclusive locks avoid the failed transactions, but also limit throughput. If each commit takes 50ms, at most 20 transactions per second are possible. Improving throughput beyond that point requires application-level changes, like batching updates.


F1 users can mix optimistic and pessimistic transactions arbitrarily and still preserve ACID semantics. All F1 writes update the last modification timestamp on every relevant lock column. Snapshot transactions are independent of any write transactions, and are also always consistent.

## Flexible Locking Granularity 灵活的锁定粒度

F1 provides row-level locking by default. Each F1 row contains one default lock column that covers all columns in the same row. However, concurrency levels can be changed in the schema. For example, users can increase concurrency by defining additional lock columns in the same row, with each lock column covering a subset of columns. In an extreme case, each column can be covered by a separate lock column, resulting in column-level locking.

One common use for column-level locking is in tables with concurrent writers, where each updates a different set of columns. For example, we could have a front-end system allowing users to change bids for keywords, and a back-end system that updates serving history on the same keywords. Busy customers may have continuous streams of bid updates at the same time that back-end systems are updating stats. Column-level locking avoids transaction conflicts between these independent streams of updates.

Users can also selectively reduce concurrency by using a lock column in a parent table to cover columns in a child table. This means that a set of rows in the child table share the same lock column and writes within this set of rows get serialized. Frequently, F1 users use lock columns in parent tables to avoid insertion phantoms for specific predicates or make other business logic constraints easier to enforce. For example, there could be a limit on keyword count per Ad-Group, and a rule that keywords must be distinct. Such constraints are easy to enforce correctly if concurrent keyword insertions (in the same AdGroup) are impossible.

# change history

Many database users build mechanisms to log changes, either from application code or using database features like triggers. In the MySQL system that AdWords used before F1, our Java application libraries added change history records into all transactions. This was nice, but it was inefficient and never 100% reliable. Some classes of changes would not get history records, including changes written from Python scripts and manual SQL data changes.

In F1, Change History is a first-class feature at the database level, where we can implement it most efficiently and can guarantee full coverage. In a change-tracked database, all tables are change-tracked by default, although specific tables or columns can be opted out in the schema. Every transaction in F1 creates one or more ChangeBatch Protocol Buffers, which include the primary key and before and after values of changed columns for each updated row. These ChangeBatches are written into normal F1 tables that exist as children of each root table. The primary key of the ChangeBatch table includes the associated root table key and the transaction commit timestamp. When a transaction updates data under multiple root rows, possibly from different root table hierarchies, one ChangeBatch is written for each distinct root row (and these ChangeBatches include pointers to each other so the full transaction can be re-assembled if necessary). This means that for each root row, the change history table includes ChangeBatches showing all changes associated with children of that root row, in commit order, and this data is easily queryable with SQL. This clustering also means that change history is stored close to the data being tracked, so these additional writes normally do not add additional participants into Spanner transactions, and therefore have minimal latency impact.


F1’s ChangeHistory mechanism has a variety of uses. The most common use is in applications that want to be notified of changes and then do some incremental processing. For example, the approval system needs to be notified when new ads have been inserted so it can approve them. F1 uses a publish-and-subscribe system to push notifications that particular root rows have changed. The publish happens in Spanner and is guaranteed to happen at least once after any series of changes to any root row. Subscribers normally remember a checkpoint (i.e. a high-water mark) for each root row and read all changes newer than the checkpoint whenever they receive a notification. This is a good example of a place where Spanner’s timestamp ordering properties are very powerful since they allow using checkpoints to guarantee that every change is processed exactly once. A separate system exists that makes it easy for these clients to see only changes to tables or columns they care about.


Change History also gets used in interesting ways for caching. One client uses an in-memory cache based on database state, distributed across multiple servers, and uses this while rendering pages in the AdWords web UI. After a user commits an update, it is important that the next page rendered reflects that update. When this client reads from the cache, it passes in the root row key and the commit timestamp of the last write that must be visible. If the cache is behind that timestamp, it reads Change History records beyond its checkpoint and applies those changes to its in-memory state to catch up. This is much cheaper than reloading the cache with a full extraction and much simpler and more accurate than comparable cache invalidation protocols.

# client design

## simplified ORM

The nature of working with a distributed data store required us to rethink the way our client applications interacted with the database. Many of our client applications had been written using a MySQL-based ORM layer that could not be adapted to work well with F1. Code written using this library exhibited several common ORM anti-patterns:

* Obscuring database operations from developers.
* Serial reads, including for loops that do one query per iteration.
* Implicit traversals: adding unwanted joins and loading unnecessary data “just in case”.

* 模糊/隐藏开发人员的数据库操作
* 串行读取，包括每次迭代执行一次查询的循环。
* 隐式遍历：添加不需要的连接并加载不必要的数据“以防万一”。

Patterns like these are common in ORM libraries. They may save development time in small-scale systems with local data stores, but they hurt scalability even then. When combined with a high-latency remote database like F1, they are disastrous. For F1, we replaced this ORM layer with a new, stripped-down API that forcibly avoids these anti-patterns. The new ORM layer does not use any joins and does not implicitly traverse any relationships between records. All object loading is explicit, and the ORM layer exposes APIs that promote the use of parallel and asynchronous read access. This is practical in an F1 schema for two reasons. First, there are simply fewer tables, and clients are usually loading Protocol Buffers directly from the database. Second, hierarchically structured primary keys make loading all children of an object expressible as a single range read without a join.

With this new F1 ORM layer, application code is more explicit and can be slightly more complex than code using the MySQL ORM, but this complexity is partially offset by the reduced impedance mismatch provided by Protocol Buffer columns. The transition usually results in better client code that uses more efficient access patterns to the database. Avoiding serial reads and other anti-patterns re- sults in code that scales better with larger data sets and exhibits a flatter overall latency distribution.

使用这个新的F1 ORM层，应用程序代码更加明确，并且可能比使用MySQL ORM的代码稍微复杂一些，但这种复杂性部分地被协议缓冲区列提供的减少的阻抗不匹配所抵消。 转换通常会产生更好的客户端代码，使用更高效的数据库访问模式。 避免串行读取和其他反模式会导致代码在更大的数据集中更好地扩展，并显示更平坦的整体延迟分布。

With MySQL, latency in our main interactive application was highly variable. Average latency was typically 200-300 ms. Small operations on small customers would run much faster than that, but large operations on large customers could be much slower, with a latency tail of requests taking multiple seconds. Developers regularly fought to identify and fix cases in their code causing excessively serial reads and high latency. With our new coding style on the F1 ORM, this doesn’t happen. User requests typically require a fixed number (fewer than 10) of reads, independent of request size or data size. The minimum latency is higher than in MySQL because of higher minimum read cost, but the average is about the same, and the latency tail for huge requests is only a few times slower than the median.

## nosql interface

F1 supports a NoSQL key/value based interface that allows for fast and simple programmatic access to rows. Read requests can include any set of tables, requesting specific columns and key ranges for each. Write requests specify inserts, updates, and deletes by primary key, with any new column values, for any set of tables.

This interface is used by the ORM layer under the hood, and is also available for clients to use directly. This API allows for batched retrieval of rows from multiple tables in a single call, minimizing the number of round trips required to complete a database transaction. Many applications prefer to use this NoSQL interface because it’s simpler to construct structured read and write requests in code than it is to generate SQL. This interface can be also be used in MapReduces to specify which data to read.


## sql interface

F1 also provides a full-fledged SQL interface, which is used for low-latency OLTP queries, large OLAP queries, and everything in between. F1 supports joining data from its Spanner data store with other data sources including Bigtable, CSV files, and the aggregated analytical data warehouse for AdWords. The SQL dialect extends standard SQL with constructs that allow accessing data stored in Protocol Buffers. Updates are also supported using SQL data manipulation statements, with extensions to support updating fields inside protocol buffers and to deal with repeated structure inside protocol buffers. Full syntax details are beyond the scope of this paper.

# query processing

The F1 SQL query processing system has the following key properties which we will elaborate on in this section:

* Queries are executed either as low-latency centrally executed queries or distributed queries with high parallelism.

* All data is remote and batching is used heavily to mitigate network latency.

* All input data and internal data is arbitrarily partitioned and has few useful ordering properties.

* Queries use many hash-based repartitioning steps.

* Individual query plan operators are designed to stream data to later operators as soon as possible, maximizing pipelining in query plans.

* Hierarchically clustered tables have optimized access methods.
* Query data can be consumed in parallel.
* Protocol Buffer-valued columns provide first-class support for structured data types.
* Spanner’s snapshot consistency model provides globally consistent results.


* 查询作为低延迟集中执行查询或具有高并行性的分布式查询执行。

* 所有数据都是远程的，并且大量使用批处理来缓解网络延迟。

* 所有输入数据和内部数据都是任意分区的，并且几乎没有有用的排序属性。

* 查询使用许多基于散列的重新分区步骤。

* 单个查询计划运算符旨在尽快将数据流式传输到后期运算符，从而最大限度地提高查询计划中的流水线。

* 分层聚簇表具有优化的访问方法。
* 查询数据可以并行使用。
* Protocol Buffer-valued列提供对结构化数据类型的一流支持。
* Spanner的快照一致性模型提供全局一致的结果。

## central and distributed Queries

F1 SQL supports both centralized and distributed execution of queries. Centralized execution is used for short OLTP-style queries and the entire query runs on one F1 server node. Distributed execution is used for OLAP-style queries and spreads the query workload over worker tasks in the F1 slave pool (see Section 2). Distributed queries always use snapshot transactions. The query optimizer uses heuristics to determine which execution mode is appropriate for a given query. In the sections that follow, we will mainly focus our attention on distributed query execution. Many of the concepts apply equally to centrally executed queries.

## distrbuted query example

The following example query would be answered using distributed execution:

```SQL
SELECT agcr.CampaignId, click.Region,
       cr.Language, SUM(click.Clicks)
FROM AdClick click
  JOIN AdGroupCreative agcr
    USING (AdGroupId, CreativeId)
  JOIN Creative cr
USING (CustomerId, CreativeId) WHERE click.Date = '2013-03-23'
GROUP BY agcr.CampaignId, click.Region, cr.Language
```

This query uses part of the AdWords schema. An AdGroup is a collection of ads with some shared configuration. A Creative is the actual ad text. The AdGroupCreative table is a link table between AdGroup and Creative; Creatives can be shared by multiple AdGroups. Each AdClick records the Creative that the user was shown and the AdGroup from which the Creative was chosen. This query takes all AdClicks on a specific date, finds the corresponding AdGroupCreative and then the Creative. It then aggregates to find the number of clicks grouped by campaign, region and language.

A possible query plan for this query is shown in Figure 3. In the query plan, data is streamed bottom-up through each of the operators up until the aggregation operator. The deepest operator performs a scan of the AdClick table. In the same worker node, the data from the AdClick scan flows into a lookup join operator, which looks up AdGroupCreative records using a secondary index key. The plan then repartitions the data stream by a hash of the CustomerId and CreativeId, and performs a lookup in a hash table that is partitioned in the same way (a distributed hash join). After the distributed hash join, the data is once again repartitioned, this time by a hash of the CampaignId, Region and Language fields, and then fed into an aggregation operator that groups by those same fields (a distributed aggregation).

## remote data

SQL query processing, and join processing in particular, poses some interesting challenges in F1, primarily because F1 does not store its data locally. F1’s main data store is Spanner, which is a remote data source, and F1 SQL can also access other remote data sources and join across them. These remote data accesses involve highly variable network latency [9]. In contrast, traditional database systems generally perform processing on the same machine that hosts their data, and they mostly optimize to reduce the number of disk seeks and disk accesses.

Network latency and disk latency are fundamentally different in two ways. First, network latency can be mitigated by batching or pipelining data accesses. F1 uses extensive batching to mitigate network latency. Secondly, disk latency is generally caused by contention for a single limited resource, the actual disk hardware. This severely limits the usefulness of sending out multiple data accesses at the same time. In contrast, F1’s network based storage is typically distributed over many disks, because Spanner partitions its data across many physical servers, and also at a finer-grained level because Spanner stores its data in CFS. This makes it much less likely that multiple data accesses will contend for the same resources, so scheduling multiple data accesses in parallel often results in near-linear speedup until the underlying storage system is truly overloaded.


The prime example of how F1 SQL takes advantage of batching is found in the lookup join query plan operator. This operator executes a join by reading from the inner table using equality lookup keys. It first retrieves rows from the outer table, extracting the lookup key values from them and deduplicating those keys. This continues until it has gathered 50MB worth of data or 100,000 unique lookup key values. Then it performs a simultaneous lookup of all keys in the inner table. This returns the requested data in arbitrary order. The lookup join operator joins the retrieved inner table rows to the outer table rows which are stored in memory, using a hash table for fast lookup. The results are streamed immediately as output from the lookup join node.

F1’s query operators are designed to stream data as much as possible, reducing the incidence of pipeline stalls. This design decision limits operators’ ability to preserve interesting data orders. Specifically, an F1 operator often has many reads running asynchronously in parallel, and streams rows to the next operator as soon as they are available. This emphasis on data streaming means that ordering properties of the input data are lost while allowing for maximum read request concurrency and limiting the space needed for row buffering.

## distrbuted execution overview

The structure of a distributed query plan is as follows. A full query plan consists of potentially tens of plan parts, each of which represents a number of workers that execute the same query subplan. The plan parts are organized as a directed acyclic graph (DAG), with data flowing up from the leaves of the DAG to a single root node, which is the only node with no out edges, i.e. the only sink. The root node, also called the query coordinator, is executed by the server that received the incoming SQL query request from a client. The query coordinator plans the query for execution, receives results from the penultimate plan parts, performs any final aggregation, sorting, or filtering, and then streams the results back to the client, except in the case of partitioned consumers as described in Section 8.6.

A technique frequently used by distributed database systems is to take advantage of an explicit co-partitioning of the stored data. Such co-partitioning can be used to push down large amounts of query processing onto each of the processing nodes that host the partitions. F1 cannot take advantage of such partitioning, in part because the data is always remote, but more importantly, because Spanner applies an arbitrary, effectively random partitioning. Moreover, Spanner can also dynamically change the partitioning. Hence, to perform operations efficiently, F1 must frequently resort to repartitioning the data. Because none of the input data is range partitioned, and because range partitioning depends on correct statistics, we have eschewed range partitioning altogether and opted to only apply hash partitioning.


Traditionally, repartitioning like this has been regarded as something to be avoided because of the heavy network traffic involved. Recent advances in the scalability of network switch hardware have allowed us to connect clusters of several hundred F1 worker processes in such a way that all servers can simultaneously communicate with each other at close to full network interface speed. This allows us to repartition without worrying much about network capacity and concepts like rack affinity. A potential downside of this solution is that it limits the size of an F1 cluster to the limits of available network switch hardware. This has not posed a problem in practice for the queries and data sizes that the F1 system deals with.


The use of hash partitioning allows us to implement an efficient distributed hash join operator and a distributed aggregation operator. These operators were already demon-strated in the example query in Section 8.2. The hash join operator repartitions both of its inputs by applying a hash function to the join keys. In the example query, the hash join keys are CustomerId and CreativeId. Each worker is responsible for a single partition of the hash join. Each worker loads its smallest input (as estimated by the query planner) into an in-memory hash table. It then reads its largest input and probes the hash table for each row, streaming out the results. For distributed aggregation, we aggregate as much as possible locally inside small buffers, then repartition the data by a hash of the grouping keys, and finally perform a full aggregation on each of the hash partitions. When hash tables grow too large to fit in memory, we apply standard algorithms that spill parts of the hash table to disk.

F1 SQL operators execute in memory, without checkpointing to disk, and stream data as much as possible. This avoids the cost of saving intermediate results to disk, so queries run as fast as the data can be processed. This means, however, that any server failure can make an entire query fail. Queries that fail get retried transparently, which usually hides these failures. In practice, queries that run for up to an hour are sufficiently reliable, but queries much longer than that may experience too many failures. We are exploring adding checkpointing for some intermediate results into our query plans, but this is challenging to do without hurting latency in the normal case where no failures occur.

## hierachical table joins

As described in Section 3.1, the F1 data model supports hierarchically clustered tables, where the rows of a child table are interleaved in the parent table. This data model allows us to efficiently join a parent table and a descendant table by their shared primary key prefix. For instance, consider the join of table Customer with table Campaign:

```SQL
SELECT *
  FROM Customer JOIN
       Campaign USING (CustomerId)
```

The hierarchically clustered data model allows F1 to perform this join using a single request to Spanner in which we request the data from both tables. Spanner will return the data to F1 in interleaved order (a pre-order depth-first traversal), ordered by primary key prefix, e.g.:

```
Customer(3)
    Campaign(3,5)
    Campaign(3,6)
  Customer(4)
    Campaign(4,2)
    Campaign(4,4)
```

While reading this stream, F1 uses a merge-join-like algorithm which we call cluster join. The cluster join operator only needs to buffer one row from each table, and returns the joined results in a streaming fashion as the Spanner input data is received. Any number of tables can be cluster joined this way using a single Spanner request, as long as all tables fall on a single ancestry path in the table hierarchy. For instance, in the following table hierarchy, F1 SQL can only join RootTable to either ChildTable1 or ChildTable2 in this way, but not both:

```
RootTable
    ChildTable1
    ChildTable2
```

When F1 SQL has to join between sibling tables like these, it will perform one join operation using the cluster join algorithm, and select an alternate join algorithm for the remaining join. An algorithm such as lookup join is able to perform this join without disk spilling or unbounded memory usage because it can construct the join result piecemeal using bounded-size batches of lookup keys.

## partitioned consumers

F1 queries can produce vast amounts of data, and pushing this data through a single query coordinator can be a bottleneck. Furthermore, a single client process receiving all the data can also be a bottleneck and likely cannot keep up with many F1 servers producing result rows in parallel. To solve this, F1 allows multiple client processes to consume sharded streams of data from the same query in parallel. This feature is used for partitioned consumers like MapReduces[10]. The client application sends the query to F1 and requests distributed data retrieval. F1 then returns a set of endpoints to connect to. The client must connect to all of these endpoints and retrieve the data in parallel. Due to the streaming nature of F1 queries, and the cross-dependencies caused by frequent hash repartitioning, slowness in one distributed reader may slow other distributed readers as well, as the F1 query produces results for all readers in lock-step. A possible, but so far unimplemented mitigation strategy for this horizontal dependency is to use disk-backed buffering to break the dependency and to allow clients to proceed independently.

## queries with Protocol Buffers

As explained in Section 3, the F1 data model makes heavy use of Protocol Buffer valued columns. The F1 SQL dialect treats these values as first class objects, providing full access to all of the data contained therein. For example, the following query requests the `CustomerId` and the entire Protocol Buffer valued column `Info` for each customer whose country code is `US`.

```SQL
SELECT c.CustomerId, c.Info
FROM Customer AS c
WHERE c.Info.country_code = 'US'
```

This query illustrates two aspects of Protocol Buffer support. First, queries use path expressions to extract individual fields (`c.Info.country code`). Second, F1 SQL also allows for querying and passing around entire protocol buffers (`c.Info`). Support for full Protocol Buffers reduces the impedance mismatch between F1 SQL and client applications, which often prefer to receive complete Protocol Buffers.

Protocol Buffers also allow `repeated fields`, which may have zero or more instances, i.e., they can be regarded as variable-length arrays. When these repeated fields occur in F1 database columns, they are actually very similar to hierarchical child tables in a 1:N relationship. The main difference between a child table and a repeated field is that the child table contains an explicit foreign key to its parent table, while the repeated field has an implicit foreign key to the Protocol Buffer containing it. Capitalizing on this similarity, F1 SQL supports access to repeated fields using PROTO JOIN, a JOIN variant that joins by the implicit foreign key. For instance, suppose that we have a table `Customer`, which has a Protocol Buffer column `Whitelist` which in turn contains a repeated field feature. Furthermore, suppose that the values of this field feature are themselves Protocol Buffers, each of which represents the whitelisting status of a particular feature for the parent `Customer`.

```SQL
SELECT c.CustomerId, f.feature
  FROM Customer AS c
PROTO JOIN c.Whitelist.feature AS f WHERE f.status = 'STATUS_ENABLED'
```

This query joins the `Customer` table with its virtual child table `Whitelist.feature` by the foreign key that is implied by containment. It then filters the resulting combinations by the value of a field `f.status` inside the child table f, and returns another field `f.feature` from that child table. In this query syntax, the PROTO JOIN specifies the parent relation of the repeated field by qualifying the repeated field name with `c`, which is the alias of the parent relation. The implementation of the PROTO JOIN construct is straightforward: in the read from the outer relation we retrieve the entire Protocol Buffer column containing the repeated field, and then for each outer row we simply enumerate the repeated field instances in memory and join them to the outer row.

F1 SQL also allows subqueries on repeated fields in Protocol Buffers. The following query has a scalar subquery to count the number of `Whitelist.features`, and an EXISTS subquery to select only `Customers` that have at least one feature that is not `ENABLED`. Each subquery iterates over repeated field values contained inside Protocol Buffers from the current row.

```SQL
SELECT c.CustomerId, c.Info,
    (SELECT COUNT(*) FROM c.Whitelist.feature) nf
  FROM Customer AS c
  WHERE EXISTS (SELECT * FROM c.Whitelist.feature f
WHERE f.status != 'ENABLED')
```

Protocol Buffers have performance implications for query processing. First, we always have to fetch entire Protocol Buffer columns from Spanner, even when we are only interested in a small subset of fields. This takes both additional network and disk bandwidth. Second, in order to extract the fields that the query refers to, we always have to parse the contents of the Protocol Buffer fields. Even though we have implemented an optimized parser to extract only requested fields, the impact of this decoding step is significant. Future versions will improve this by pushing parsing and field selection to Spanner, thus reducing network bandwidth required and saving CPU in F1 while possibly using more CPU in Spanner.

# deployment

The F1 and Spanner clusters currently deployed for AdWords use five datacenters spread out across mainland US. The Spanner configuration uses 5-way Paxos replication to ensure high availability. Each region has additional readonly replicas that do not participate in the Paxos algorithm. Read-only replicas are used only for snapshot reads and thus allow us to segregate OLTP and OLAP workloads.

副本数 5
Intuitively, 3-way replication should suffice for high availability. In practice, this is not enough. When one datacenter is down (because of either an outage or planned maintenance), both surviving replicas must remain available for F1 to be able to commit transactions, because a Paxos commit must succeed on a majority of replicas. If a second datacenter goes down, the entire database becomes completely unavailable. Even a single machine failure or restart temporarily removes a second replica, causing unavailability for data hosted on that server.

Spanner’s Paxos implementation designates one of the replicas as a leader. All transactional reads and commits must be routed to the leader replica. User transactions normally require at least two round trips to the leader (reads followed by a commit). Moreover, F1 servers usually perform an extra read as a part of transaction commit (to get old values for Change History, index updates, optimistic transaction timestamp verification, and referential integrity checks). Consequently, transaction latency is best when clients and F1 servers are co-located with Spanner leader replicas. We designate one of the datacenters as a preferred leader location. Spanner locates leader replicas in the preferred leader location whenever possible. Clients that perform heavy database modifications are usually deployed close to the preferred leader location. Other clients, including those that primarily run queries, can be deployed anywhere, and normally do reads against local F1 servers.

写尽量在replica leader

We have chosen to deploy our five read/write replicas with two each on the east and west coasts of the US, and the fifth centrally. With leaders on the east coast, commits require round trips to the other east coast datacenter, plus the central datacenter, which accounts for the 50ms minimum latency. We have chosen this deployment to maximize availability in the presence of large regional outages. Other F1 and Spanner instances could be deployed with closer replicas to reduce commit latency.

# latency and throughput

In our configuration, F1 users see read latencies of 5-10 ms and commit latencies of 50-150 ms. Commit latency is largely determined by network latency between datacenters. The Paxos algorithm allows a transaction to commit once a majority of voting Paxos replicas acknowledge the transaction. With five replicas, commits require a round trip from the leader to the two nearest replicas. Multi-group commits require 2PC, which typically doubles the minimum latency.

Despite the higher database latency, overall user-facing latency for the main interactive AdWords web application averages about 200ms, which is similar to the preceding system running on MySQL. Our schema clustering and application coding strategies have successfully hidden the inherent latency of synchronous commits. Avoiding serial reads in client code accounts for much of that. In fact, while the average is similar, the MySQL application exhibited tail latency much worse than the same application on F1.

For non-interactive applications that apply bulk updates, we optimize for throughput rather than latency. We typically structure such applications so they do small transactions, scoped to single Spanner directories when possible, and use parallelism to achieve high throughput. For example, we have one application that updates billions of rows per day, and we designed it to perform single-directory transactions of up to 500 rows each, running in parallel and aiming for 500 transactions per second. F1 and Spanner provide very high throughput for parallel writes like this and are usually not a bottleneck – our rate limits are usually chosen to protect downstream Change History consumers who can’t process changes fast enough.

For query processing, we have mostly focused on functionality and parity so far, and not on absolute query performance. Small central queries reliably run in less than 10ms, and some applications do tens or hundreds of thousands of SQL queries per second. Large distributed queries run with latency comparable to MySQL. Most of the largest queries actually run faster in F1 because they can use more parallelism than MySQL, which can only parallelize up to the number of MySQL shards. In F1, such queries often see linear speedup when given more resources.


Resource costs are usually higher in F1, where queries often use an order of magnitude more CPU than similar MySQL queries. MySQL stored data uncompressed on local disk, and was usually bottlenecked by disk rather than CPU, even with flash disks. F1 queries start from data compressed on disk and go through several layers, decompressing, pro- cessing, recompressing, and sending over the network, all of which have significant cost. Improving CPU efficiency here is an area for future work.

# related work

As a hybrid of relational and NoSQL systems, F1 is related to work in both areas. F1’s relational query execution techniques are similar to those described in the shared-nothing database literature, e.g., [12], with some key differences like the ignoring of interesting orders and the absence of copartitioned data. F1’s NoSQL capabilities share properties with other well-described scalable key-value stores including Bigtable [6], HBase [1], and Dynamo [11]. The hierarchical schema and clustering properties are similar to Megastore [3].

Optimistic transactions have been used in previous systems including Percolator [19] and Megastore [3]. There is extensive literature on transaction, consistency and locking models, including optimistic and pessimistic transactions, such as [24] and [4].


Prior work [15] also chose to mitigate the inherent latency of remote lookups though the use of asynchrony in query processing. However, due to the large volumes of data processed by F1, our system is not able to make the simplifying assumption that an unlimited number of asynchronous requests can be made at the same time. This complication, coupled with the high variability of storage operation latency, led to the out-of-order streaming design described in Section 8.3.

MDCC [17] suggests some Paxos optimizations that could be applied to reduce the overhead of multi-participant transactions.

Using Protocol Buffers as first class types makes F1, in part, a kind of object database [2]. The resulting simplified ORM results in a lower impedance mismatch for most client applications at Google, where Protocol Buffers are used pervasively. The hierarchical schema and clustering properties are similar to Megastore and ElasTraS [8]. F1 treats repeated fields inside protocol buffers like nested relations [21].

# conclusion

In recent years, conventional wisdom in the engineering community has been that if you need a highly scalable, high-throughput data store, the only viable option is to use a NoSQL key/value store, and to work around the lack of ACID transactional guarantees and the lack of conveniences like secondary indexes, SQL, and so on. When we sought a replacement for Google’s MySQL data store for the AdWords product, that option was simply not feasible: the complexity of dealing with a non-ACID data store in every part of our business logic would be too great, and there was simply no way our business could function without SQL queries. Instead of going NoSQL, we built F1, a distributed relational database system that combines high availability, the throughput and scalability of NoSQL systems, and the functionality, usability and consistency of traditional relational databases, including ACID transactions and SQL queries.

Google’s core AdWords business is now running completely on F1. F1 provides the SQL database functionality that our developers are used to and our business requires. Unlike our MySQL solution, F1 is trivial to scale up by simply adding machines. Our low-level commit latency is higher, but by using a coarse schema design with rich column types and improving our client application coding style, the observable end-user latency is as good as before and the worst-case latencies have actually improved.

F1 shows that it is actually possible to have a highly scalable and highly available distributed database that still provides all of the guarantees and conveniences of a traditional relational database.


# acknowledgements

We would like to thank the Spanner team, without whose great efforts we could not have built F1. We’d also like to thank the many developers and users across all AdWords teams who migrated their systems to F1, and who played a large role influencing and validating the design of this system. We also thank all former and newer F1 team members, including Michael Armbrust who helped write this paper, and Marcel Kornacker who worked on the early design of the F1 query engine.
