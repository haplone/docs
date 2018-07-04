# About
This document is an updated version of the original design documents
by Spencer Kimball from early 2014.

本文档是由Spencer Kimball从2014年初开始编写的原始设计文档的更新版本。


# Overview

CockroachDB is a distributed SQL database. The primary design goals
are **scalability**, **strong consistency** and **survivability**
(hence the name). CockroachDB aims to tolerate disk, machine, rack, and
even **datacenter failures** with minimal latency disruption and **no
manual intervention**. CockroachDB nodes are symmetric; a design goal is
**homogeneous deployment** (one binary) with minimal configuration and
no required external dependencies.

CockroachDB是一个分布式SQL数据库。 主要设计目标是**可扩展性**，**强一致性**和**生存能力**（由此得名）。 蟑螂数据库旨在容忍磁盘，机器，机架和甚至是**数据中心故障**，延迟时间最短，而且不需要人工干预**。CockroachDB节点是对称的; 设计目标是**同构部署**（一个二进制），最小配置，没有需要的外部依赖。

The entry point for database clients is the SQL interface. Every node
in a CockroachDB cluster can act as a client SQL gateway. A SQL
gateway transforms and executes client SQL statements to key-value
(KV) operations, which the gateway distributes across the cluster as
necessary and returns results to the client. CockroachDB implements a
**single, monolithic sorted map** from key to value where both keys
and values are byte strings.

数据库客户端的入口点是SQL接口。 CockroachDB集群中的每个节点都可以充当客户端SQL网关(gateway)。 一个SQL网关(gateway)将客户端SQL语句转换并执行key-value（KV）操作，并根据需要在集群中分配这些操作,将结果返回给客户。 CockroachDB实现了一个**单一，整体排序的映射**从key到value映射，而且都是字节字符串。

The KV map is logically composed of smaller segments of the keyspace called
ranges. Each range is backed by data stored in a local KV storage engine (we
use [RocksDB](http://rocksdb.org/), a variant of
[LevelDB](https://github.com/google/leveldb)). Range data is replicated to a
configurable number of additional CockroachDB nodes. Ranges are merged and
split to maintain a target size, by default `64M`. The relatively small size
facilitates quick repair and rebalancing to address node failures, new capacity
and even read/write load. However, the size must be balanced against the
pressure on the system from having more ranges to manage.

KV映射在逻辑上由称为ranges的keyspace的较小片段组成。 每个range都由存储在本地KV存储引擎 [RocksDB](http://rocksdb.org/)中，这是一个类似 [LevelDB](https://github.com/google/leveldb)的实现。 range数据被复制到一定数量（可配置）的CockroachDB节点。 ranges合并和分割以保持目标大小，默认为`64M`。 相对较小的尺寸有助于快速修复和重新平衡以解决节点故障和新容量问题甚至读写负载。 但是，尺寸必须与其平衡系统受到更多rangs的管理压力。

CockroachDB achieves horizontally scalability:
- adding more nodes increases the capacity of the cluster by the
  amount of storage on each node (divided by a configurable
  replication factor), theoretically up to 4 exabytes (4E) of logical
  data;
- client queries can be sent to any node in the cluster, and queries
  can operate independently (w/o conflicts), meaning that overall
  throughput is a linear factor of the number of nodes in the cluster.
- queries are distributed (ref: distributed SQL) so that the overall
  throughput of single queries can be increased by adding more nodes.

CockroachDB实现了横向扩展性：
- 通过添加更多节点可增加群集的容量。每个节点上的存储量（除以可配置的复制因子replication factor），理论上高达4艾字节（4E）的逻辑数据;
- 客户端查询可以发送到集群中的任何节点，并可以独立查询（没有冲突），这意味着整体吞吐量与群集中节点数量的线性关系。
- 查询分发（参考：分布式SQL），通过增加更多的节点来增加整体查询的吞吐量。

CockroachDB achieves strong consistency:
- uses a distributed consensus protocol for synchronous replication of
  data in each key value range. We’ve chosen to use the [Raft
  consensus algorithm](https://raftconsensus.github.io); all consensus
  state is stored in RocksDB.
- single or batched mutations to a single range are mediated via the
  range's Raft instance. Raft guarantees ACID semantics.
- logical mutations which affect multiple ranges employ distributed
  transactions for ACID semantics. CockroachDB uses an efficient
  **non-locking distributed commit** protocol.

CockroachDB实现了强一致性：
- 同步range中数据副本时使用分布式一致性协议。我们选择使用[raft共识算法]（https://raftconsensus.github.io）; 所有的共识状态存储在RocksDB中。
- 对range的单个或批量修改通过range的raft实例实现。Raft保证ACID语义。
- 影响多个range的逻辑修改采用ACID语义的分布式事务。 CockroachDB使用高效的**non-locking distributed commit**协议。


CockroachDB achieves survivability:
- range replicas can be co-located within a single datacenter for low
  latency replication and survive disk or machine failures. They can
  be distributed across racks to survive some network switch failures.
- range replicas can be located in datacenters spanning increasingly
  disparate geographies to survive ever-greater failure scenarios from
  datacenter power or networking loss to regional power failures
  (e.g. `{ US-East-1a, US-East-1b, US-East-1c }`, `{ US-East, US-West,
  Japan }`, `{ Ireland, US-East, US-West}`, `{ Ireland, US-East,
  US-West, Japan, Australia }`).

蟑螂数据库实现了高可用：
- range的副本放在同一个数据中心，可以实现低延迟复制，容忍磁盘和机器故障。他们能分布在机架上以容忍一些网络交换机故障。
- range的副本可以放在多个的数据中心，可以容忍更多的失败场景，像数据中心电力或网络损失、区域性电力故障
   （例如， `{ US-East-1a, US-East-1b, US-East-1c }`, `{ US-East, US-West,
   Japan }`, `{ Ireland, US-East, US-West}`, `{ Ireland, US-East,
   US-West, Japan, Australia }`）。


CockroachDB provides [snapshot
isolation](http://en.wikipedia.org/wiki/Snapshot_isolation) (SI) and
serializable snapshot isolation (SSI) semantics, allowing **externally
consistent, lock-free reads and writes**--both from a historical snapshot
timestamp and from the current wall clock time. SI provides lock-free reads
and writes but still allows write skew. SSI eliminates write skew, but
introduces a performance hit in the case of a contentious system. SSI is the
default isolation; clients must consciously decide to trade correctness for
performance. CockroachDB implements [a limited form of linearizability
](#strict-serializability-linearizability), providing ordering for any
observer or chain of observers.

CockroachDB提供[快照隔离snapshot isolation](http://en.wikipedia.org/wiki/Snapshot_isolation)（SI）和 可串行快照隔离serializable snapshot isolation（SSI）语义，允许外部使用**一致的，无锁读取和写入** -- 都来自历史快照时间戳和当前挂钟时间。 SI提供无锁读取并写入但仍允许写入歪斜write skew。 SSI消除了写入歪斜，但是在引起争议的系统中引入性能问题。 SSI是默认隔离机制; 客户端必须地明确指定以正确性换取
性能。 CockroachDB实现了[线性化的一种有限形式]（＃strict-serializability-linearizability），为任何提供排序观察员或观察员链。

Similar to
[Spanner](http://static.googleusercontent.com/media/research.google.com/en/us/archive/spanner-osdi2012.pdf)
directories, CockroachDB allows configuration of arbitrary zones of data.
This allows replication factor, storage device type, and/or datacenter
location to be chosen to optimize performance and/or availability.
Unlike Spanner, zones are monolithic and don’t allow movement of fine
grained data on the level of entity groups.

如同[spanner](http://static.googleusercontent.com/media/research.google.com/en/us/archive/spanner-osdi2012.pdf)目录，CockroachDB允许配置任意数据zones。这允许复制因子，存储设备类型和/或数据中心选择位置以优化性能和/或可用性。与spanner不同，zones是单片的，不允许移动有关实体组织级别的数据。

# Architecture

CockroachDB implements a layered architecture. The highest level of
abstraction is the SQL layer (currently unspecified in this document).
It depends directly on the [*SQL layer*](#sql),
which provides familiar relational concepts
such as schemas, tables, columns, and indexes. The SQL layer
in turn depends on the [distributed key value store](#key-value-api),
which handles the details of range addressing to provide the abstraction
of a single, monolithic key value store. The distributed KV store
communicates with any number of physical cockroach nodes. Each node
contains one or more stores, one per physical device.

CockroachDB实现了分层架构。 最高级别的抽象是SQL层（在本文档中没有细说）。熟悉的关系概念如模式，表格，列和索引直接取决于[* SQL layer*](＃sql)。 SQL层依次依赖[分布式key value存储]()＃key-value-api)，它处理范围寻址的细节以提供抽象一个单一的key value存储store。 分布式KV存储与任何数量的物理CockroachDB节点进行通信。 每个节点包含一个或多个store，每个物理设备一个。


![Architecture](media/architecture.png)

Each store contains potentially many ranges, the lowest-level unit of
key-value data. Ranges are replicated using the Raft consensus protocol.
The diagram below is a blown up version of stores from four of the five
nodes in the previous diagram. Each range is replicated three ways using
raft. The color coding shows associated range replicas.

每个store可以有很多个range，key value数据的最低层单位。 range使用Raft共识协议复制。下面的图表显示上面图中五个store中四个store中节点的细节。 每个range使用raft复制3份。 颜色代码显示相关的range副本。

![Ranges](media/ranges.png)

Each physical node exports two RPC-based key value APIs: one for
external clients and one for internal clients (exposing sensitive
operational features). Both services accept batches of requests and
return batches of responses. Nodes are symmetric in capabilities and
exported interfaces; each has the same binary and may assume any
role.

每个物理节点暴露两个基于RPC的key value API：一个用于外部客户端和一个内部客户端（暴露敏感操作功能）。 这两个服务都可以接受批量的请求并返回一批响应。 节点在功能和暴露接口方面是一样的; 每个都有相同的二进制，并可能会假设任何角色。

Nodes and the ranges they provide access to can be arranged with various
physical network topologies to make trade offs between reliability and
performance. For example, a triplicated (3-way replica) range could have
each replica located on different:

-   disks within a server to tolerate disk failures.
-   servers within a rack to tolerate server failures.
-   servers on different racks within a datacenter to tolerate rack power/network failures.
-   servers in different datacenters to tolerate large scale network or power outages.

nodes和ranges可以通过不同的网络拓扑访问，以平衡可靠性和性能。 例如，一个三重复制（3路复制）range可能有
位于不同的位置副本：

- 服务器内的磁盘，容忍磁盘故障。
- 机架内的服务器，可以容忍服务器故障。
- 数据中心内不同机架上的服务器，可以容忍机架电源/网络故障。
- 位于不同数据中心的服务器，可以容忍大规模网络或停电。

Up to `F` failures can be tolerated, where the total number of replicas `N = 2F + 1` (e.g. with 3x replication, one failure can be tolerated; with 5x replication, two failures, and so on).

在'N = 2F + 1'的副本总数中（例如3x复制，一个故障可以容忍; 5x复制，两个故障等等），可以容忍高达'F'个副本故障。

# Keys

Cockroach keys are arbitrary byte arrays. Keys come in two flavors:
system keys and table data keys. System keys are used by Cockroach for
internal data structures and metadata. Table data keys contain SQL
table data (as well as index data). System and table data keys are
prefixed in such a way that all system keys sort before any table data
keys.

cockroach 的 key 是任意字节数组。 key有两种类型：系统key和表数据key。 系统key由Cockroach用于内部数据结构和元数据。 表数据key包含SQL表数据（以及索引数据）。 系统和表格数据key都有前缀，以保障系统key会排序在表数据key之前。

System keys come in several subtypes:

- **Global** keys store cluster-wide data such as the "meta1" and
    "meta2" keys as well as various other system-wide keys such as the
    node and store ID allocators.
- **Store local** keys are used for unreplicated store metadata
    (e.g. the `StoreIdent` structure). "Unreplicated" indicates that
    these values are not replicated across multiple stores because the
    data they hold is tied to the lifetime of the store they are
    present on.
- **Range local** keys store range metadata that is associated with a
    global key. Range local keys have a special prefix followed by a
    global key and a special suffix. For example, transaction records
    are range local keys which look like:
    `\x01k<global-key>txn-<txnID>`.
- **Replicated Range ID local** keys store range metadata that is
    present on all of the replicas for a range. These keys are updated
    via Raft operations. Examples include the range lease state and
    abort cache entries.
- **Unreplicated Range ID local** keys store range metadata that is
    local to a replica. The primary examples of such keys are the Raft
    state and Raft log.

- **Gloal** 键存储集群范围的数据，如“meta1”和“meta2”key以及各种其他系统级key，例如节点和商店ID分配器。
- **store local** key用于未复制的store元数据（例如`StoreIdent`结构）。 “Unreplicated”表示这些值不会跨多个store复制，因为他们持有的数据与目前store的生命周期相关。
- **range local** 存储range元数据的key与global key相关。range local keys有一个特殊的前缀，后跟一个global key和特殊后缀。例如，事物记录是range local key，如下所示：`\ x01k <全局密钥> txn- <txnID>`。
- **Replicated Range ID local** 全部副本上都有的存储range 元数据的key 在range。这些key通过Raft操作更新。例子包括range租约状态和中止缓存条目。
- **Unreplicated Range ID local** 只在本地副本有的存储range元数据的key。这种eky的主要例子是raft状态和raft日志。

Table data keys are used to store all SQL data. Table data keys
contain internal structure as described in the section on [mapping
data between the SQL model and
KV](#data-mapping-between-the-sql-model-and-kv).

表数据key用于存储所有SQL数据。 表数据key包含内部结构在[SQL模型与KV映射数据 mapping data between the SQL model and KV](#data-mapping-between-the-sql-model-and-kv)中介绍。


# Versioned Values

Cockroach maintains historical versions of values by storing them with
associated commit timestamps. Reads and scans can specify a snapshot
time to return the most recent writes prior to the snapshot timestamp.
Older versions of values are garbage collected by the system during
compaction according to a user-specified expiration interval. In order
to support long-running scans (e.g. for MapReduce), all versions have a
minimum expiration.

Versioned values are supported via modifications to RocksDB to record
commit timestamps and GC expirations per key.

Cockroach 通过关联提交时间戳方式实现value多个历史版本的存储。读取和扫描可以指定一个快照时间来返回快照时间戳之前的最新写入。根据用户指定的到期间隔，系统在压缩compaction时，垃圾回收value旧版本数据。为了支持长时间运行的扫描（例如MapReduce），所有版本都有一个最小超时时间。


通过修改RocksDB来记录每个key的提交时间戳和GC超时时间，来实现value的多版本。

# Lock-Free Distributed Transactions

Cockroach provides distributed transactions without locks. Cockroach
transactions support two isolation levels:

Cockroach 无锁方式实现分布式事务。Cockroach事务支持2种隔离级别：

- snapshot isolation (SI) and
- *serializable* snapshot isolation (SSI).

*SI* is simple to implement, highly performant, and correct for all but a
handful of anomalous conditions (e.g. write skew). *SSI* requires just a touch
more complexity, is still highly performant (less so with contention), and has
no anomalous conditions. Cockroach’s SSI implementation is based on ideas from
the literature and some possibly novel insights.

*SI* 实施简单，性能高，除了少数异常情况（例如写歪斜），都是正确的。 *SSI* 只需要多一点点复杂度，仍然是高性能的（争用更少），并且没有异常情况。 Cockroach的SSI实现基于literature和一些可能的新颖见解。

SSI is the default level, with SI provided for application developers
who are certain enough of their need for performance and the absence of
write skew conditions to consciously elect to use it. In a lightly
contended system, our implementation of SSI is just as performant as SI,
requiring no locking or additional writes. With contention, our
implementation of SSI still requires no locking, but will end up
aborting more transactions. Cockroach’s SI and SSI implementations
prevent starvation scenarios even for arbitrarily long transactions.

SSI是默认级别，如果应用程序开发人员明确性能需求，并且没有write skew情况，可以使用SI级别。在轻度竞争系统中，我们的SSI实现没有锁和额外写入，具有跟SI一样的性能。在有竞争的情况，我们的SSI实现一样没有锁，但是会有更多的事务被取消。Cockroach的SI和SSI实现，可以防止任意长事务导致的饥饿。


See the [Cahill paper](https://drive.google.com/file/d/0B9GCVTp_FHJIcEVyZVdDWEpYYXVVbFVDWElrYUV0NHFhU2Fv/edit?usp=sharing)
for one possible implementation of SSI. This is another [great paper](http://cs.yale.edu/homes/thomson/publications/calvin-sigmod12.pdf).
For a discussion of SSI implemented by preventing read-write conflicts
(in contrast to detecting them, called write-snapshot isolation), see
the [Yabandeh paper](https://drive.google.com/file/d/0B9GCVTp_FHJIMjJ2U2t6aGpHLTFUVHFnMTRUbnBwc2pLa1RN/edit?usp=sharing),
which is the source of much inspiration for Cockroach’s SSI.

请参阅[Cahill论文]（https://drive.google.com/file/d/0B9GCVTp_FHJIcEVyZVdDWEpYYXVVbFVDWElrYUV0NHFhU2Fv/edit?usp=sharing）为SSI的一个可能的实施。 这是另一个[伟大的论文]（http://cs.yale.edu/homes/thomson/publications/calvin-sigmod12.pdf）。有关通过防止读写冲突实施的SSI的讨论
（与检测它们不同，称为写入快照隔离），请参阅[Yabandeh论文]（https://drive.google.com/file/d/0B9GCVTp_FHJIMjJ2U2t6aGpHLTFUVHFnMTRUbnBwc2pLa1RN/edit?usp=sharing），这是蟑螂SSI的灵感源泉


Both SI and SSI require that the outcome of reads must be preserved, i.e.
a write of a key at a lower timestamp than a previous read must not succeed. To
this end, each range maintains a bounded *in-memory* cache from key range to
the latest timestamp at which it was read.

Most updates to this *timestamp cache* correspond to keys being read, though
the timestamp cache also protects the outcome of some writes (notably range
deletions) which consequently must also populate the cache. The cache’s entries
are evicted oldest timestamp first, updating the low water mark of the cache
appropriately.

SI和SSI都要求必须保留读取的结果，即
写入比以前的读取更低的时间戳的密钥一定不会成功。 至
为此，每个范围在关键范围内维护一个有界的*内存*缓存
它读取的最新时间戳。

虽然这个*时间戳高速缓存的大部分更新对应于被读取的键
时间戳缓存也保护一些写入的结果（特别是范围
删除），因此也必须填充缓存。 缓存的条目
首先被逐出最老的时间戳，更新缓存的低水位标记
适当。

Each Cockroach transaction is assigned a random priority and a
"candidate timestamp" at start. The candidate timestamp is the
provisional timestamp at which the transaction will commit, and is
chosen as the current clock time of the node coordinating the
transaction. This means that a transaction without conflicts will
usually commit with a timestamp that, in absolute time, precedes the
actual work done by that transaction.

In the course of coordinating a transaction between one or more
distributed nodes, the candidate timestamp may be increased, but will
never be decreased. The core difference between the two isolation levels
SI and SSI is that the former allows the transaction's candidate
timestamp to increase and the latter does not.

每个蟑螂交易分配一个随机优先级和一个
“候选时间戳”在开始。 候选时间戳是
临时时间戳记，交易将在其中进行，并且是
选择为协调节点的当前时钟时间
交易。 这意味着一个没有冲突的交易将会
通常会提供一个在绝对时间之前的时间戳
该交易完成的实际工作。

在协调一个或多个交易的过程中
分布式节点，候选时间戳可能会增加，但会
永远不会减少。 两个隔离级别的核心区别
SI和SSI是前者允许交易的候选人
时间戳增加，后者不会

**Hybrid Logical Clock**

Each cockroach node maintains a hybrid logical clock (HLC) as discussed
in the [Hybrid Logical Clock paper](http://www.cse.buffalo.edu/tech-reports/2014-04.pdf).
HLC time uses timestamps which are composed of a physical component (thought of
as and always close to local wall time) and a logical component (used to
distinguish between events with the same physical component). It allows us to
track causality for related events similar to vector clocks, but with less
overhead. In practice, it works much like other logical clocks: When events
are received by a node, it informs the local HLC about the timestamp supplied
with the event by the sender, and when events are sent a timestamp generated by
the local HLC is attached.

如所讨论的，每个蟑螂节点保持混合逻辑时钟（HLC）
在[混合逻辑时钟文件]（http://www.cse.buffalo.edu/tech-reports/2014-04.pdf）中。
HLC时间使用由物理组件组成的时间戳（想到
作为和总是接近本地墙上时间）和一个逻辑组件（用于
区分具有相同物理组件的事件）。它使我们能够
跟踪类似于矢量时钟的相关事件的因果关系，但具有较少的因果关系
高架。实际上，它与其他逻辑时钟非常相似：当发生事件时
由节点接收，它通知本地HLC提供的时间戳
由发件人发送事件，并在发送事件时生成一个时间戳
本地HLC已连接。

For a more in depth description of HLC please read the paper. Our
implementation is [here](https://github.com/cockroachdb/cockroach/blob/master/pkg/util/hlc/hlc.go).

Cockroach picks a Timestamp for a transaction using HLC time. Throughout this
document, *timestamp* always refers to the HLC time which is a singleton
on each node. The HLC is updated by every read/write event on the node, and
the HLC time >= wall time. A read/write timestamp received in a cockroach request
from another node is not only used to version the operation, but also updates
the HLC on the node. This is useful in guaranteeing that all data read/written
on a node is at a timestamp < next HLC time.



有关HLC的更深入的描述，请阅读文章。我们的
实现是[here]（https://github.com/cockroachdb/cockroach/blob/master/pkg/util/hlc/hlc.go）。

蟑螂使用HLC时间为交易选择时间戳。纵观这一点
文件，*时间戳*总是指HLC时间，这是一个单身人士
在每个节点上。 HLC由节点上的每个读/写事件更新，并且
HLC时间> =挂墙时间。在蟑螂请求中收到的读取/写入时间戳
从另一个节点不仅用于版本操作，而且还更新
节点上的HLC。这在保证读取/写入所有数据方面很有用
在节点上的时间标记<下一个HLC时间。

**Transaction execution flow**

Transactions are executed in two phases:

1. Start the transaction by selecting a range which is likely to be
   heavily involved in the transaction and writing a new transaction
   record to a reserved area of that range with state "PENDING". In
   parallel write an "intent" value for each datum being written as part
   of the transaction. These are normal MVCC values, with the addition of
   a special flag (i.e. “intent”) indicating that the value may be
   committed after the transaction itself commits. In addition,
   the transaction id (unique and chosen at txn start time by client)
   is stored with intent values. The txn id is used to refer to the
   transaction record when there are conflicts and to make
   tie-breaking decisions on ordering between identical timestamps.
   Each node returns the timestamp used for the write (which is the
   original candidate timestamp in the absence of read/write conflicts);
   the client selects the maximum from amongst all write timestamps as the
   final commit timestamp.

 1.通过选择可能的范围来启动交易
    大量参与交易并撰写新的交易
    记录到状态为“PENDING”的该范围的保留区域。在
    并行地为写入的每个数据写入“intent”值作为其一部分
    的交易。这些都是正常的MVCC值，加上
    表示该值可能是的特殊标志（即“意图”）
    在交易本身提交后承诺。此外，
    交易ID（唯一并在客户端txn开始时选择）
    与意图值一起存储。 txn id用于引用
    交易记录何时有冲突并作出
    打破决定相同时间戳之间的顺序。
    每个节点返回用于写入的时间戳（这是
    在没有读/写冲突的情况下的原始候选时间戳）;
    客户端从所有写入时间戳中选择最大值作为
    最终提交时间戳。

2. Commit the transaction by updating its transaction record. The value
   of the commit entry contains the candidate timestamp (increased as
   necessary to accommodate any latest read timestamps). Note that the
   transaction is considered fully committed at this point and control
   may be returned to the client.

   In the case of an SI transaction, a commit timestamp which was
   increased to accommodate concurrent readers is perfectly
   acceptable and the commit may continue. For SSI transactions,
   however, a gap between candidate and commit timestamps
   necessitates transaction restart (note: restart is different than
   abort--see below).

   After the transaction is committed, all written intents are upgraded
   in parallel by removing the “intent” flag. The transaction is
   considered fully committed before this step and does not wait for
   it to return control to the transaction coordinator.


 2.通过更新交易记录来确认交易。价值
    提交条目中包含候选时间戳（增加为
    以适应任何最新的读取时间戳）。请注意
    交易在这一点和控制下被视为完全承诺
    可能会退还给客户。

    在SI事务的情况下，提交时间戳是
    增加以适应并发读者是完美的
    可以接受并且承诺可能会继续。对于SSI交易，
    然而，候选和提交时间戳之间存在差距
    有必要重新启动交易（注意：重新启动不同于
    中止 - 见下文）。

    交易完成后，所有书面意向都会升级
    同时删除“意图”标志。交易是
    在此步骤之前考虑完全承诺并且不会等待
    它将控制权交还给交易协调员。

In the absence of conflicts, this is the end. Nothing else is necessary
to ensure the correctness of the system.




在没有冲突的情况下，这是结局。没有别的是必要的
以确保系统的正确性。

**Conflict Resolution**

Things get more interesting when a reader or writer encounters an intent
record or newly-committed value in a location that it needs to read or
write. This is a conflict, usually causing either of the transactions to
abort or restart depending on the type of conflict.

当读者或作家遇到意图时，事情会变得更有趣
在需要阅读的地点记录或重新提交价值
写。 这是一种冲突，通常会导致任何一项交易
根据冲突类型中止或重新启动。

***Transaction restart:***

This is the usual (and more efficient) type of behaviour and is used
except when the transaction was aborted (for instance by another
transaction).
In effect, that reduces to two cases; the first being the one outlined
above: An SSI transaction that finds upon attempting to commit that
its commit timestamp has been pushed. The second case involves a transaction
actively encountering a conflict, that is, one of its readers or writers
encounter data that necessitate conflict resolution
(see transaction interactions below).

When a transaction restarts, it changes its priority and/or moves its
timestamp forward depending on data tied to the conflict, and
begins anew reusing the same txn id. The prior run of the transaction might
have written some write intents, which need to be deleted before the
transaction commits, so as to not be included as part of the transaction.
These stale write intent deletions are done during the reexecution of the
transaction, either implicitly, through writing new intents to
the same keys as part of the reexecution of the transaction, or explicitly,
by cleaning up stale intents that are not part of the reexecution of the
transaction. Since most transactions will end up writing to the same keys,
the explicit cleanup run just before committing the transaction is usually
a NOOP.


这是通常（和更有效）的行为类型，并且被使用
除非交易被中止（例如通过另一次交易）
交易）。
实际上，这减少到两种情况;第一个是概述的
上面：一个SSI事务，在尝试提交时发现
其提交时间戳已被推送。第二种情况涉及交易
积极地遇到冲突，即其读者或作者之一
遇到需要解决冲突的数据
（请参阅下面的交易互动）。

当事务重新启动时，它会改变其优先级和/或移动其优先级
时间戳转发取决于与冲突相关的数据，以及
开始重新使用相同的txn id。事务的先前运行可能会
写了一些写意图，需要先删除
交易提交，以便不被纳入交易。
这些陈旧的写意图删除在重新执行期间完成
交易，或者通过写新的意图来暗示
作为重新执行交易的一部分，或明确地说，
通过清理不属于重新执行的陈旧意图
交易。由于大多数交易最终会写入相同的密钥，
通常在提交事务之前运行显式清理
NOOP。


***Transaction abort:***

This is the case in which a transaction, upon reading its transaction
record, finds that it has been aborted. In this case, the transaction
can not reuse its intents; it returns control to the client before
cleaning them up (other readers and writers would clean up dangling
intents as they encounter them) but will make an effort to clean up
after itself. The next attempt (if applicable) then runs as a new
transaction with **a new txn id**.


这是交易在阅读交易时的情况
记录，发现它已被中止。 在这种情况下，交易
不能重用它的意图; 它将控制权返回给客户端
清理他们（其他读者和作家会清理悬挂
意图，因为他们遇到他们），但会努力清理
本身之后。 下一次尝试（如果适用）然后作为新的运行
交易**新的txn id **。


***Transaction interactions:***

There are several scenarios in which transactions interact:

事务交互有几种情况：

- **Reader encounters write intent or value with newer timestamp far
  enough in the future**: This is not a conflict. The reader is free
  to proceed; after all, it will be reading an older version of the
  value and so does not conflict. Recall that the write intent may
  be committed with a later timestamp than its candidate; it will
  never commit with an earlier one. **Side note**: if a SI transaction
  reader finds an intent with a newer timestamp which the reader’s own
  transaction has written, the reader always returns that intent's value.
- **阅读器遇到写意图或价值较远的新时间戳
未来足够**：这不是冲突。读者是免费的
继续;毕竟，它会读取旧版本的
价值等都不冲突。回想一下写意图可能
比其候选人的时间戳晚;它会
从来没有提前做过。 **附注**：如果是SI交易
读者找到了读者自己拥有的更新时间戳的意图
事务已经写入，读者总是返回该意图的价值。

- **Reader encounters write intent or value with newer timestamp in the
  near future:** In this case, we have to be careful. The newer
  intent may, in absolute terms, have happened in our read's past if
  the clock of the writer is ahead of the node serving the values.
  In that case, we would need to take this value into account, but
  we just don't know. Hence the transaction restarts, using instead
  a future timestamp (but remembering a maximum timestamp used to
  limit the uncertainty window to the maximum clock skew). In fact,
  this is optimized further; see the details under "choosing a time
  stamp" below.



- **阅读器遇到写意图或价值与更新的时间戳在
 不久的将来：**在这种情况下，我们必须小心。更新
 在绝对意义上，意图可能发生在我们阅读过去的情况下
 作者的时钟在提供值的节点之前。
 在这种情况下，我们需要考虑这个值，但是
 我们只是不知道。因此事务重新启动，而是使用
 未来的时间戳记（但记住过去的最大时间戳记
 将不确定性窗口限制为最大时钟偏差）。事实上，
 这进一步优化;请参阅“选择时间”下的详细信息
 邮票“。

- **Reader encounters write intent with older timestamp**: the reader
  must follow the intent’s transaction id to the transaction record.
  If the transaction has already been committed, then the reader can
  just read the value. If the write transaction has not yet been
  committed, then the reader has two options. If the write conflict
  is from an SI transaction, the reader can *push that transaction's
  commit timestamp into the future* (and consequently not have to
  read it). This is simple to do: the reader just updates the
  transaction’s commit timestamp to indicate that when/if the
  transaction does commit, it should use a timestamp *at least* as
  high. However, if the write conflict is from an SSI transaction,
  the reader must compare priorities. If the reader has the higher priority,
  it pushes the transaction’s commit timestamp (that
  transaction will then notice its timestamp has been pushed, and
  restart). If it has the lower or same priority, it retries itself using as
  a new priority `max(new random priority, conflicting txn’s
  priority - 1)`.

- **阅读器遇到用旧时间戳写入的意图**：阅读器
 必须遵循意图的交易ID到交易记录。
 如果交易已经被提交，那么读者可以
 只是读取价值。如果写入事务尚未完成
 承诺，那么读者有两种选择。如果写入冲突
 来自SI交易，读者可以*推动该交易
 将时间戳投入未来*（因此不必
 阅读）。这很容易做：读者只是更新
 事务的提交时间戳以指示何时/如果
 事务确实提交，它应该使用时间戳*至少* as
 高。但是，如果写入冲突来自SSI事务，
 读者必须比较优先事项。如果读者具有更高的优先级，
 它推送事务的提交时间戳（即
 交易会注意到它的时间戳已经被推送了
 重新开始）。如果它具有较低或相同的优先级，则使用as重试自身
 一个新的优先级`max（新的随机优先级，冲突的txn's
 优先级 - 1）`。

- **Writer encounters uncommitted write intent**:
  If the other write intent has been written by a transaction with a lower
  priority, the writer aborts the conflicting transaction. If the write
  intent has a higher or equal priority the transaction retries, using as a new
  priority *max(new random priority, conflicting txn’s priority - 1)*;
  the retry occurs after a short, randomized backoff interval.



- ** Writer遇到未提交的写入意图**：
 如果另一个写入意图已经由较低的事务写入
 优先级，作者放弃冲突交易。如果写
 意图具有更高或同等优先级的交易重试，作为新的使用
 优先级* max（新随机优先级，冲突txn的优先级 - 1）*;
 重试发生在短暂的随机退避间隔之后。

- **Writer encounters newer committed value**:
  The committed value could also be an unresolved write intent made by a
  transaction that has already committed. The transaction restarts. On restart,
  the same priority is reused, but the candidate timestamp is moved forward
  to the encountered value's timestamp.

- ** Writer遇到更新的承诺值**：
  承诺的价值也可能是一个未解决的写作意图
  已经提交的交易。 交易重新开始。 重启时，
  重用相同的优先级，但候选时间戳会向前移动
  到遇到的值的时间戳。

- **Writer encounters more recently read key**:
  The *read timestamp cache* is consulted on each write at a node. If the write’s
  candidate timestamp is earlier than the low water mark on the cache itself
  (i.e. its last evicted timestamp) or if the key being written has a read
  timestamp later than the write’s candidate timestamp, this later timestamp
  value is returned with the write. A new timestamp forces a transaction
  restart only if it is serializable.



- ** Writer遇到最近读取的键**：
  读取时间戳高速缓存*在每个节点上写入时查阅。 如果写的
  候选时间戳早于缓存本身的低水位标记
  （即其最后驱逐的时间戳）或者正被写入的密钥是否具有读取
  时间戳晚于写入的候选时间戳，这是稍后的时间戳
  值与写入一起返回。 新的时间戳会强制执行事务
  仅在序列化时才重新启动。

**Transaction management**

Transactions are managed by the client proxy (or gateway in SQL Azure
parlance). Unlike in Spanner, writes are not buffered but are sent
directly to all implicated ranges. This allows the transaction to abort
quickly if it encounters a write conflict. The client proxy keeps track
of all written keys in order to resolve write intents asynchronously upon
transaction completion. If a transaction commits successfully, all intents
are upgraded to committed. In the event a transaction is aborted, all written
intents are deleted. The client proxy doesn’t guarantee it will resolve intents.

事务由客户端代理（或SQL Azure中的网关）管理
说法）。与Spanner不同，写入不会被缓冲，而是被发送
直接涉及所有涉及的范围。这允许交易中止
如果遇到写入冲突，很快。客户端代理保持跟踪
所有书面密钥，以便在异步时解决写意图
交易完成。如果一个事务成功提交，所有的意图
升级为承诺。在交易中止的情况下，全部书面
意图被删除。客户端代理不保证它将解析意图。

In the event the client proxy restarts before the pending transaction is
committed, the dangling transaction would continue to "live" until
aborted by another transaction. Transactions periodically heartbeat
their transaction record to maintain liveness.
Transactions encountered by readers or writers with dangling intents
which haven’t been heartbeat within the required interval are aborted.
In the event the proxy restarts after a transaction commits but before
the asynchronous resolution is complete, the dangling intents are upgraded
when encountered by future readers and writers and the system does
not depend on their timely resolution for correctness.



如果客户端代理在挂起的事务处理之前重新启动
承诺，悬而未决的交易将继续“活”直到
被另一笔交易中止。交易周期性心跳
他们的交易记录保持活跃。
读者或作家悬而未决的意图遇到的交易
在要求的时间间隔内没有心跳的心跳被中止。
如果代理在事务提交后但之前重新启动
异步解析完成后，悬挂意图升级
当被未来的读者和作家和系统遇到时
不依赖于他们及时解决正确性。

An exploration of retries with contention and abort times with abandoned
transaction is
[here](https://docs.google.com/document/d/1kBCu4sdGAnvLqpT-_2vaTbomNmX3_saayWEGYu1j7mQ/edit?usp=sharing).

探索与争议和废弃时间中止的重试
交易是
[这里]（https://docs.google.com/document/d/1kBCu4sdGAnvLqpT-_2vaTbomNmX3_saayWEGYu1j7mQ/edit?usp=sharing）

**Transaction Records**

Please see [pkg/roachpb/data.proto](https://github.com/cockroachdb/cockroach/blob/master/pkg/roachpb/data.proto) for the up-to-date structures, the best entry point being `message Transaction`.

请参阅[pkg / roachpb / data.proto]（https://github.com/cockroachdb/cockroach/blob/master/pkg/roachpb/data.proto）了解最新的结构，最好的入口点是 `消息交易`。

**Pros**

- No requirement for reliable code execution to prevent stalled 2PC
  protocol.
- Readers never block with SI semantics; with SSI semantics, they may
  abort.
- Lower latency than traditional 2PC commit protocol (w/o contention)
  because second phase requires only a single write to the
  transaction record instead of a synchronous round to all
  transaction participants.
- Priorities avoid starvation for arbitrarily long transactions and
  always pick a winner from between contending transactions (no
  mutual aborts).
- Writes not buffered at client; writes fail fast.
- No read-locking overhead required for *serializable* SI (in contrast
  to other SSI implementations).
- Well-chosen (i.e. less random) priorities can flexibly give
  probabilistic guarantees on latency for arbitrary transactions
  (for example: make OLTP transactions 10x less likely to abort than
  low priority transactions, such as asynchronously scheduled jobs).

- 不需要可靠的代码执行来防止停滞的2PC
  协议。
- 读者不会阻止SI语义; 与SSI语义，他们可能
 中止。
- 比传统的2PC提交协议更低的延迟（无争用）
  因为第二阶段只需要一次写入
  交易记录而不是所有的同步轮
  交易参与者。
- 重点避免了任意长时间的交易和饥饿
  总是从竞争交易中选择一个胜利者（没有
  相互放弃）。
- 不在客户端缓冲的写入; 写入失败很快。
- 对于* serializable * SI，不需要读取锁定开销（相反，
  到其他SSI实现）。
- 精心挑选（即较少随机）的优先事项可以灵活地给予
  随机事务延迟的概率保证
  （例如：使OLTP交易的可能性降低10倍
  低优先级事务，例如异步调度的作业）。

**Cons**

- Reads from non-lease holder replicas still require a ping to the lease holder
  to update the *read timestamp cache*.
- Abandoned transactions may block contending writers for up to the
  heartbeat interval, though average wait is likely to be
  considerably shorter (see [graph in link](https://docs.google.com/document/d/1kBCu4sdGAnvLqpT-_2vaTbomNmX3_saayWEGYu1j7mQ/edit?usp=sharing)).
  This is likely considerably more performant than detecting and
  restarting 2PC in order to release read and write locks.
- Behavior different than other SI implementations: no first writer
  wins, and shorter transactions do not always finish quickly.
  Element of surprise for OLTP systems may be a problematic factor.
- Aborts can decrease throughput in a contended system compared with
  two phase locking. Aborts and retries increase read and write
  traffic, increase latency and decrease throughput.

- 从非租赁持有人副本中读取仍需要ping给租赁持有人
  更新*读取时间戳缓存*。
- 被遗弃的交易可能阻止竞争作家达成协议
  心跳间隔，尽管平均等待时间可能是
  （见[链接图]（https://docs.google.com/document/d/1kBCu4sdGAnvLqpT-_2vaTbomNmX3_saayWEGYu1j7mQ/edit?usp=sharing））。
  这可能比检测更有效
  重新启动2PC以释放读取和写入锁定。
- 与其他SI实现不同的行为：没有第一个作者
  胜利和较短的交易并不总是很快完成。
  OLTP系统的惊喜元素可能是一个有问题的因素。
- 中止可以降低竞争系统的吞吐量
  两相锁定。 中止和重试增加读取和写入
  流量，增加延迟并降低吞吐量。

**Choosing a Timestamp**

A key challenge of reading data in a distributed system with clock skew
is choosing a timestamp guaranteed to be greater than the latest
timestamp of any committed transaction (in absolute time). No system can
claim consistency and fail to read already-committed data.

Accomplishing consistency for transactions (or just single operations)
accessing a single node is easy. The timestamp is assigned by the node
itself, so it is guaranteed to be at a greater timestamp than all the
existing timestamped data on the node.

For multiple nodes, the timestamp of the node coordinating the
transaction `t` is used. In addition, a maximum timestamp `t+ε` is
supplied to provide an upper bound on timestamps for already-committed
data (`ε` is the maximum clock skew). As the transaction progresses, any
data read which have timestamps greater than `t` but less than `t+ε`
cause the transaction to abort and retry with the conflicting timestamp
t<sub>c</sub>, where t<sub>c</sub> \> t. The maximum timestamp `t+ε` remains
the same. This implies that transaction restarts due to clock uncertainty
can only happen on a time interval of length `ε`.

在时钟偏斜的分布式系统中读取数据的关键挑战
正在选择一个保证大于最新时间戳的时间戳
任何已提交交易的时间戳（绝对时间）。没有系统可以
声明一致性并且无法读取已提交的数据。

完成交易的一致性（或者只是单个操作）
访问单个节点很容易。时间戳由节点分配
本身，所以它保证比所有的时间戳更大
节点上现有的时间戳数据。

对于多个节点，协调节点的节点的时间戳
事务`t`被使用。另外，最大时间戳t +ε是
提供用于为已提交的时间戳提供上限
数据（`ε`是最大时钟偏差）。随着交易的进展，任何
数据读取的时间戳大于“t”但小于“t +ε”
导致事务中止并使用冲突的时间戳重试
t <sub> c </ sub>，其中t <sub> c </ sub> \> t。最大时间戳t +ε仍然存在
一样。这意味着事务由于时钟不确定性而重新启动
只能发生在长度为ε的时间间隔上。

We apply another optimization to reduce the restarts caused
by uncertainty. Upon restarting, the transaction not only takes
into account t<sub>c</sub>, but the timestamp of the node at the time
of the uncertain read t<sub>node</sub>. The larger of those two timestamps
t<sub>c</sub> and t<sub>node</sub> (likely equal to the latter) is used
to increase the read timestamp. Additionally, the conflicting node is
marked as “certain”. Then, for future reads to that node within the
transaction, we set `MaxTimestamp = Read Timestamp`, preventing further
uncertainty restarts.

Correctness follows from the fact that we know that at the time of the read,
there exists no version of any key on that node with a higher timestamp than
t<sub>node</sub>. Upon a restart caused by the node, if the transaction
encounters a key with a higher timestamp, it knows that in absolute time,
the value was written after t<sub>node</sub> was obtained, i.e. after the
uncertain read. Hence the transaction can move forward reading an older version
of the data (at the transaction's timestamp). This limits the time uncertainty
restarts attributed to a node to at most one. The tradeoff is that we might
pick a timestamp larger than the optimal one (> highest conflicting timestamp),
resulting in the possibility of a few more conflicts.

We expect retries will be rare, but this assumption may need to be
revisited if retries become problematic. Note that this problem does not
apply to historical reads. An alternate approach which does not require
retries makes a round to all node participants in advance and
chooses the highest reported node wall time as the timestamp. However,
knowing which nodes will be accessed in advance is difficult and
potentially limiting. Cockroach could also potentially use a global
clock (Google did this with [Percolator](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Peng.pdf)),
which would be feasible for smaller, geographically-proximate clusters.

我们应用另一个优化来减少重新启动造成的
由不确定性。重新启动后，交易不仅需要
考虑到t <sub> c </ sub>，但是当时节点的时间戳
不确定的读取t <sub>节点</ sub>。这两个时间戳中的较大者
t <sub> c </ sub>和t <sub> node </ sub>（可能等于后者）
增加读取时间戳。另外，冲突节点是
标记为“确定”。然后，为将来读取到该节点内的节点
事务中，我们设置了`MaxTimestamp = Read Timestamp`，进一步防止
不确定性重启。

正确性来自于我们知道在阅读时，
该节点上没有任何版本的密钥的时间戳高于
吨<子>节点</子>。在节点重新启动后，如果事务发生
遇到具有更高时间戳的密钥时，它知道在绝对时间内，
该值是在获得t <sub>节点</ sub>之后写入的，即在
不确定的读。因此，交易可以继续阅读旧版本
的数据（在交易的时间戳）。这限制了时间的不确定性
重新启动归因于一个节点至多一个。权衡是我们可能的
选择大于最优时间戳的时间戳（>冲突时间最高的时间戳），
导致多一些冲突的可能性。

我们预计重试将很少见，但这种假设可能需要
如果重试成为问题，则重新访问。请注意，这个问题没有
适用于历史阅读。一种不需要的替代方法
重试事先向所有节点参与者发送一个回合
选择报告的最高节点时间作为时间戳。然而，
知道哪些节点将被预先访问是困难的
潜在的限制。蟑螂也可能使用全球性的
（谷歌用[Percolator]（https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Peng.pdf））做了这个，），
这对较小的地理上接近的群集是可行的。

# Strict Serializability (Linearizability)

Roughly speaking, the gap between <i>strict serializability</i> (which we use
interchangeably with <i>linearizability</i>) and CockroachDB's default
isolation level (<i>serializable</i>) is that with linearizable transactions,
causality is preserved. That is, if one transaction (say, creating a posting
for a user) waits for its predecessor (creating the user in the first place)
to complete, one would hope that the logical timestamp assigned to the former
is larger than that of the latter.
In practice, in distributed databases this may not hold, the reason typically
being that clocks across a distributed system are not perfectly synchronized
and the "later" transaction touches a part disjoint from that on which the
first transaction ran, resulting in clocks with disjoint information to decide
on the commit timestamps.

In practice, in CockroachDB many transactional workloads are actually
linearizable, though the precise conditions are too involved to outline them
here.

Causality is typically not required for many transactions, and so it is
advantageous to pay for it only when it *is* needed. CockroachDB implements
this via <i>causality tokens</i>: When committing a transaction, a causality
token can be retrieved and passed to the next transaction, ensuring that these
two transactions get assigned increasing logical timestamps.

Additionally, as better synchronized clocks become a standard commodity offered
by cloud providers, CockroachDB can provide global linearizability by doing
much the same that [Google's
Spanner](http://research.google.com/archive/spanner.html) does: wait out the
maximum clock offset after committing, but before returning to the client.

See the blog post below for much more in-depth information.

https://www.cockroachlabs.com/blog/living-without-atomic-clocks/

粗略地说，严格的可序列化</ i>（我们使用它）之间的差距
与线性化</ i>互换）和CockroachDB的默认设置
隔离级别（<serializable </ i>）是使用线性化事务，
因果关系被保留下来。也就是说，如果一个交易（比如创建一个发布）
对于用户）等待其前任（首先创建用户）
要完成，人们会希望分配给前者的逻辑时间戳
大于后者。
实际上，在分布式数据库中这可能不成立，原因通常是
因为分布式系统中的时钟并不完全同步
而“后来的”交易触及了一部分与之不相关的部分
第一次交易运行，导致时钟与不相交的信息来决定
在提交时间戳上。

实际上，在CockroachDB中，实际上很多事务性工作负载
可线性化，虽然精确的条件太牵涉到勾勒出它们
这里。

因为许多交易通常不需要因果关系，所以它是
只有在*需要时才有利于支付。 CockroachDB实现
通过<i>因果关系标记</ i>：当提交一个事务时，一个因果关系
令牌可以被检索并传递给下一个事务，确保这些
两个事务被分配越来越多的逻辑时间戳。

另外，随着更好的同步时钟成为所提供的标准商品
通过云提供商，CockroachDB可以通过做提供全球线性化
这与[Google的
扳手]（http://research.google.com/archive/spanner.html）确实：等待
提交后的最大时钟偏移量，但在返回给客户端之前。

有关更深入的信息，请参阅下面的博客文章。

# Logical Map Content

Logically, the map contains a series of reserved system key/value
pairs preceding the actual user data (which is managed by the SQL
subsystem).

逻辑上，地图包含一系列保留的系统键/值
实际用户数据之前的对（由SQL管理）
子系统）。

- `\x02<key1>`: Range metadata for range ending `\x03<key1>`. This a "meta1" key.
- ...
- `\x02<keyN>`: Range metadata for range ending `\x03<keyN>`. This a "meta1" key.
- `\x03<key1>`: Range metadata for range ending `<key1>`. This a "meta2" key.
- ...
- `\x03<keyN>`: Range metadata for range ending `<keyN>`. This a "meta2" key.
- `\x04{desc,node,range,store}-idegen`: ID generation oracles for various component types.
- `\x04status-node-<varint encoded Store ID>`: Store runtime metadata.
- `\x04tsd<key>`: Time-series data key.
- `<key>`: A user key. In practice, these keys are managed by the SQL
  subsystem, which employs its own key anatomy.

# Stores and Storage

Nodes contain one or more stores. Each store should be placed on a unique disk.
Internally, each store contains a single instance of RocksDB with a block cache
shared amongst all of the stores in a node. And these stores in turn have
a collection of range replicas. More than one replica for a range will never
be placed on the same store or even the same node.

Early on, when a cluster is first initialized, the few default starting ranges
will only have a single replica, but as soon as other nodes are available they
will replicate to them until they've reached their desired replication factor,
the default being 3.

Zone configs can be used to control a range's replication factor and add
constraints as to where the range's replicas can be located. When there is a
change in a range's zone config, the range will up or down replicate to the
appropriate number of replicas and move its replicas to the appropriate stores
based on zone config's constraints.

节点包含一个或多个商店。 每个商店应放置在一个唯一的磁盘上。
在内部，每个商店都包含一个带有块缓存的RocksDB实例
在节点中的所有商店之间共享。 而这些商店又有
范围复制品的集合。 一个范围的多个副本将永远不会
放置在同一个商店或甚至同一个节点上。

早期，当一个集群首次被初始化时，几个默认的起始范围
将只有一个副本，但只要其他节点可用，他们
将复制到他们，直到他们已经达到他们想要的复制因子，
默认值是3。

区域配置可用于控制范围的复制因子并添加
关于范围副本可以位于何处的限制。 当有一个
更改范围的区域配置，范围将向上或向下复制到
适当数量的副本并将其副本移动到适当的存储区
基于区域配置的限制。

# Self Repair

If a store has not been heard from (gossiped their descriptors) in some time,
the default setting being 5 minutes, the cluster will consider this store to be
dead. When this happens, all ranges that have replicas on that store are
determined to be unavailable and removed. These ranges will then upreplicate
themselves to other available stores until their desired replication factor is
again met. If 50% or more of the replicas are unavailable at the same time,
there is no quorum and the whole range will be considered unavailable until at
least greater than 50% of the replicas are again available.

如果商店在一段时间内还没有听到（说出他们的描述符）
默认设置为5分钟，群集将认为该商店是
死。 发生这种情况时，所有在该存储上具有副本的范围都是
确定不可用并被删除。 这些范围将会复制
他们自己到其他可用的商店，直到他们想要的复制因素
再次遇见。 如果同时有50％或更多的副本不可用，
没有法定人数，整个范围将被视为无法使用，直到在
至少大于50％的副本可以再次使用。

# Rebalancing

As more data are added to the system, some stores may grow faster than others.
To combat this and to spread the overall load across the full cluster, replicas
will be moved between stores maintaining the desired replication factor. The
heuristics used to perform this rebalancing include:

- the number of replicas per store
- the total size of the data used per store
- free space available per store

In the future, some other factors that might be considered include:

- cpu/network load per store
- ranges that are used together often in queries
- number of active ranges per store
- number of range leases held per store

随着更多数据被添加到系统中，一些商店的增长速度可能会比其他商店快。
为了解决这个问题并将整个负载分散到整个集群中，副本
将在商店之间移动，保持所需的复制因子。该
用于执行这种再平衡的启发式算法包括：

- 每个商店的副本数量
- 每个商店使用的数据的总大小
- 每间商店的可用空间

未来，可能考虑的其他因素包括：

- 每个商店的cpu /网络负载
- 经常在查询中一起使用的范围
- 每个商店的有效范围数量
- 每个商店举行的范围租约数量

# Range Metadata

The default approximate size of a range is 64M (2\^26 B). In order to
support 1P (2\^50 B) of logical data, metadata is needed for roughly
2\^(50 - 26) = 2\^24 ranges. A reasonable upper bound on range metadata
size is roughly 256 bytes (3\*12 bytes for the triplicated node
locations and 220 bytes for the range key itself). 2\^24 ranges \* 2\^8
B would require roughly 4G (2\^32 B) to store--too much to duplicate
between machines. Our conclusion is that range metadata must be
distributed for large installations.

To keep key lookups relatively fast in the presence of distributed metadata,
we store all the top-level metadata in a single range (the first range). These
top-level metadata keys are known as *meta1* keys, and are prefixed such that
they sort to the beginning of the key space. Given the metadata size of 256
bytes given above, a single 64M range would support 64M/256B = 2\^18 ranges,
which gives a total storage of 64M \* 2\^18 = 16T. To support the 1P quoted
above, we need two levels of indirection, where the first level addresses the
second, and the second addresses user data. With two levels of indirection, we
can address 2\^(18 + 18) = 2\^36 ranges; each range addresses 2\^26 B, and
altogether we address 2\^(36+26) B = 2\^62 B = 4E of user data.

范围的默认近似大小为64M（2 \ ^ 26 B）。为了
支持1P（2 \ ^ 50 B）的逻辑数据，大致需要元数据
2×（50-26）= 2×24范围。范围元数据的合理上限
大小大约为256个字节（3 \ * 12字节为三重节点
范围密钥本身的位置和220个字节）。 2 \ ^ 24范围\ * 2 \ ^ 8
B将需要大约4G（2 \ ^ 32 B）来存储 - 太多复制
机器之间。我们的结论是范围元数据必须是
分发给大型设备。

为了在存在分布式元数据的情况下保持密钥查找速度相对较快，
我们将所有顶级元数据存储在一个范围内（第一个范围）。这些
顶级元数据键被称为* meta1 *键，并且前缀如此
他们排序到关键空间的开始。给定的元数据大小为256
以上给出的字节，单个64M范围将支持64M / 256B = 2 \ ^ 18范围，
这给出了64M的总存储量* 2 * ^ 18 = 16T。支持1P引用
在上面，我们需要两个间接级别，其中第一级解决了这个问题
第二，第二个地址用户数据。有两个层面的间接性，我们
可以解决2 \ ^（18 + 18）= 2 \ ^ 36范围;每个范围地址2 \ ^ 26 B，和
总共我们解决了用户数据的2 ^（36 + 26）B = 2 \ ^ 62 B = 4E。

For a given user-addressable `key1`, the associated *meta1* record is found
at the successor key to `key1` in the *meta1* space. Since the *meta1* space
is sparse, the successor key is defined as the next key which is present. The
*meta1* record identifies the range containing the *meta2* record, which is
found using the same process. The *meta2* record identifies the range
containing `key1`, which is again found the same way (see examples below).

Concretely, metadata keys are prefixed by `\x02` (meta1) and `\x03`
(meta2); the prefixes `\x02` and `\x03` provide for the desired
sorting behaviour. Thus, `key1`'s *meta1* record will reside at the
successor key to `\x02<key1>`.

对于给定的用户可寻址的`key1`，找到关联的* meta1 *记录
在* meta1 *空间中`key1`的后继键。 自* meta1 *空间
是稀疏的，后继键被定义为存在的下一个键。该
* meta1 *记录标识包含* meta2 *记录的范围，即
使用相同的过程找到。 * meta2 *记录标识范围
包含`key1`，这又是以同样的方式找到的（参见下面的例子）。

具体来说，元数据键以`\ x02`（meta1）和`\ x03`为前缀
（meta2）; 前缀`\ x02`和`\ x03`提供了所需的
排序行为。 因此，`key1`的* meta1 *记录将驻留在
`\ x02 <key1>`的后继键。

Note: we append the end key of each range to meta{1,2} records because
the RocksDB iterator only supports a Seek() interface which acts as a
Ceil(). Using the start key of the range would cause Seek() to find the
key *after* the meta indexing record we’re looking for, which would
result in having to back the iterator up, an option which is both less
efficient and not available in all cases.

The following example shows the directory structure for a map with
three ranges worth of data. Ellipses indicate additional key/value
pairs to fill an entire range of data. For clarity, the examples use
`meta1` and `meta2` to refer to the prefixes `\x02` and `\x03`. Except
for the fact that splitting ranges requires updates to the range
metadata with knowledge of the metadata layout, the range metadata
itself requires no special treatment or bootstrapping.

注意：我们将每个范围的结束键追加到meta {1,2}记录中，因为
RocksDB迭代器仅支持作为a的Seek（）接口
小区（）。 使用范围的开始键会导致Seek（）找到
键*后*我们正在寻找的元索引记录，这将
导致不得不支持迭代器，这是一个更少的选项
有效并且在所有情况下都不可用。

以下示例显示了具有的地图的目录结构
三个数值范围的数据。 省略号表示附加的键/值
对来填充整个范围的数据。 为了清楚起见，这些示例使用
`meta1`和`meta2`来引用前缀`\ x02`和`\ x03`。 除
因为分割范围需要更新范围
具有元数据布局知识的元数据，范围元数据
本身不需要特殊处理或自举。

**Range 0** (located on servers `dcrama1:8000`, `dcrama2:8000`,
  `dcrama3:8000`)

- `meta1\xff`: `dcrama1:8000`, `dcrama2:8000`, `dcrama3:8000`
- `meta2<lastkey0>`: `dcrama1:8000`, `dcrama2:8000`, `dcrama3:8000`
- `meta2<lastkey1>`: `dcrama4:8000`, `dcrama5:8000`, `dcrama6:8000`
- `meta2\xff`: `dcrama7:8000`, `dcrama8:8000`, `dcrama9:8000`
- ...
- `<lastkey0>`: `<lastvalue0>`

**Range 1** (located on servers `dcrama4:8000`, `dcrama5:8000`,
`dcrama6:8000`)

- ...
- `<lastkey1>`: `<lastvalue1>`

**Range 2** (located on servers `dcrama7:8000`, `dcrama8:8000`,
`dcrama9:8000`)

- ...
- `<lastkey2>`: `<lastvalue2>`

Consider a simpler example of a map containing less than a single
range of data. In this case, all range metadata and all data are
located in the same range:

考虑一个包含少于一个的地图的简单例子
数据范围。 在这种情况下，所有范围元数据和所有数据都是
位于相同的范围内：

**Range 0** (located on servers `dcrama1:8000`, `dcrama2:8000`,
`dcrama3:8000`)*

- `meta1\xff`: `dcrama1:8000`, `dcrama2:8000`, `dcrama3:8000`
- `meta2\xff`: `dcrama1:8000`, `dcrama2:8000`, `dcrama3:8000`
- `<key0>`: `<value0>`
- `...`

Finally, a map large enough to need both levels of indirection would
look like (note that instead of showing range replicas, this
example is simplified to just show range indexes):

最后，足够大的地图需要两个间接级别
看起来像（请注意，而不是显示范围副本，这
示例简化为只显示范围索引）：

**Range 0**

- `meta1<lastkeyN-1>`: Range 0
- `meta1\xff`: Range 1
- `meta2<lastkey1>`:  Range 1
- `meta2<lastkey2>`:  Range 2
- `meta2<lastkey3>`:  Range 3
- ...
- `meta2<lastkeyN-1>`: Range 262143

**Range 1**

- `meta2<lastkeyN>`: Range 262144
- `meta2<lastkeyN+1>`: Range 262145
- ...
- `meta2\xff`: Range 500,000
- ...
- `<lastkey1>`: `<lastvalue1>`

**Range 2**

- ...
- `<lastkey2>`: `<lastvalue2>`

**Range 3**

- ...
- `<lastkey3>`: `<lastvalue3>`

**Range 262144**

- ...
- `<lastkeyN>`: `<lastvalueN>`

**Range 262145**

- ...
- `<lastkeyN+1>`: `<lastvalueN+1>`

Note that the choice of range `262144` is just an approximation. The
actual number of ranges addressable via a single metadata range is
dependent on the size of the keys. If efforts are made to keep key sizes
small, the total number of addressable ranges would increase and vice
versa.

From the examples above it’s clear that key location lookups require at
most three reads to get the value for `<key>`:

请注意，范围“262144”的选择只是一个近似值。该
实际可通过单个元数据范围寻址的范围数量为
取决于键的大小。 如果努力保持关键尺寸
很小，可寻址范围的总数将会增加并且成为副作用
反之亦然。

从上面的例子可以清楚地看出，关键位置查找需要
最多三次读取以获得`<key>`的值：

1. lower bound of `meta1<key>`
2. lower bound of `meta2<key>`,
3. `<key>`.

For small maps, the entire lookup is satisfied in a single RPC to Range 0. Maps
containing less than 16T of data would require two lookups. Clients cache both
levels of range metadata, and we expect that data locality for individual
clients will be high. Clients may end up with stale cache entries. If on a
lookup, the range consulted does not match the client’s expectations, the
client evicts the stale entries and possibly does a new lookup.

对于小型地图，整个查找在单个RPC中满足范围0.地图
包含少于16T的数据将需要两次查找。 客户端缓存两个
范围元数据的级别，我们期望个人的数据局部性
客户会很高。 客户端可能会以失效的缓存条目结束。 如果在一个
查询，咨询的范围与客户的期望不符，
客户端驱逐陈旧的条目，并可能做一个新的查找。

# Raft - Consistency of Range Replicas

Each range is configured to consist of three or more replicas, as specified by
their ZoneConfig. The replicas in a range maintain their own instance of a
distributed consensus algorithm. We use the [*Raft consensus algorithm*](https://raftconsensus.github.io)
as it is simpler to reason about and includes a reference implementation
covering important details.
[ePaxos](https://www.cs.cmu.edu/~dga/papers/epaxos-sosp2013.pdf) has
promising performance characteristics for WAN-distributed replicas, but
it does not guarantee a consistent ordering between replicas.

Raft elects a relatively long-lived leader which must be involved to
propose commands. It heartbeats followers periodically and keeps their logs
replicated. In the absence of heartbeats, followers become candidates
after randomized election timeouts and proceed to hold new leader
elections. Cockroach weights random timeouts such that the replicas with
shorter round trip times to peers are more likely to hold elections
first (not implemented yet). Only the Raft leader may propose commands;
followers will simply relay commands to the last known leader.

Our Raft implementation was developed together with CoreOS, but adds an extra
layer of optimization to account for the fact that a single Node may have
millions of consensus groups (one for each Range). Areas of optimization
are chiefly coalesced heartbeats (so that the number of nodes dictates the
number of heartbeats as opposed to the much larger number of ranges) and
batch processing of requests.
Future optimizations may include two-phase elections and quiescent ranges
(i.e. stopping traffic completely for inactive ranges).

每个范围都配置为由三个或更多副本组成，如所指定的
他们的ZoneConfig。范围内的副本维护自己的实例
分布式共识算法。我们使用[* Raft一致性算法*]（https://raftconsensus.github.io）
因为推理和包含参考实现更简单
涵盖重要细节。
[ePaxos]（https://www.cs.cmu.edu/~dga/papers/epaxos-sosp2013.pdf）有
对于广域网分布式复制品来说，前景看好，但是
它不保证副本之间的一致顺序。

拉夫特选出一位相对长期的领导者，必须参与其中
提出命令。它定期心跳追随者并保留其日志
复制。在没有心跳的情况下，追随者成为候选人
随机选举超时后，继续举行新领导
选举。蟑螂权衡随机超时，使副本与
往返同行的短途往返时间更有可能举行选举
首先（尚未实施）。只有筏领导可以提出命令;
追随者只会将命令传递给最后一位已知的领导者。

我们的Raft实现与CoreOS一起开发，但增加了额外的功能
优化层来说明单个节点可能具有的事实
数百万个共识组（每个范围一个）。优化领域
主要是合并的心跳（所以节点的数量决定了）
心跳次数而不是更多的范围）和
批处理请求。
未来的优化可能包括两阶段选举和静止的范围
（即完全停止非活动范围的交通）。

# Range Leases

As outlined in the Raft section, the replicas of a Range are organized as a
Raft group and execute commands from their shared commit log. Going through
Raft is an expensive operation though, and there are tasks which should only be
carried out by a single replica at a time (as opposed to all of them).
In particular, it is desirable to serve authoritative reads from a single
Replica (ideally from more than one, but that is far more difficult).

For these reasons, Cockroach introduces the concept of **Range Leases**:
This is a lease held for a slice of (database, i.e. hybrid logical) time.
A replica establishes itself as owning the lease on a range by committing
a special lease acquisition log entry through raft. The log entry contains
the replica node's epoch from the node liveness table--a system
table containing an epoch and an expiration time for each node. A node is
responsible for continuously updating the expiration time for its entry
in the liveness table. Once the lease has been committed through raft
the replica becomes the lease holder as soon as it applies the lease
acquisition command, guaranteeing that when it uses the lease it has
already applied all prior writes on the replica and can see them locally.

如Raft部分所述，范围的副本被组织为a
Raft组并执行共享提交日志中的命令。经历
虽然筏是一个昂贵的操作，并且有任务应该只是
一次只进行一次复制（而不是全部复制）。
特别是，需要从单一的服务中提供权威的阅读
复制品（理想情况下来自多个，但这是困难得多）。

由于这些原因，蟑螂介绍** Range Leases **的概念：
这是针对一段（数据库，即混合逻辑）时间持有的租约。
复制品通过承诺确定自己拥有一定范围内的租约
通过木筏获得特殊租赁获取日志条目。日志条目包含
复制节点的节点活跃表的纪元 - 一个系统
表包含每个节点的历元和过期时间。节点是
负责不断更新其条目的到期时间
在活力表中。一旦租约通过木筏实施
该复制品一旦适用租赁即成为租赁持有人
获取命令，保证当它使用它的租约时
已经将所有先前的写入应用于副本并可以在本地看到它们。

To prevent two nodes from acquiring the lease, the requestor includes a copy
of the lease that it believes to be valid at the time it requests the lease.
If that lease is still valid when the new lease is applied, it is granted,
or another lease is granted in the interim and the requested lease is
ignored. A lease can move from node A to node B only after node A's
liveness record has expired and its epoch has been incremented.

Note: range leases for ranges within the node liveness table keyspace and
all ranges that precede it, including meta1 and meta2, are not managed using
the above mechanism to prevent circular dependencies.

A replica holding a lease at a specific epoch can use the lease as long as
the node epoch hasn't changed and the expiration time hasn't passed.
The replica holding the lease may satisfy reads locally, without incurring the
overhead of going through Raft, and is in charge or involved in handling
Range-specific maintenance tasks such as splitting, merging and rebalancing

为了防止两个节点获得租约，请求者包含一个副本
在它要求租约时它认为有效的租约。
如果在新的租约适用时该租约仍然有效，
或者在此期间批准另一项租赁，并且所要求的租约是
忽略。租约只能在节点A之后从节点A移动到节点B.
活性记录已经过期，其历元已增加。

注意：节点活跃表密钥空间和范围内的范围租约
它之前的所有范围，包括meta1和meta2，都不使用管理
上述防止循环依赖的机制。

在特定时期持有租约的副本可以使用租约
节点历元没有改变，到期时间还没有过去。
持有租约的副本可能会满足本地读取，而不会导致该问题
经过Raft的开销，并负责或涉及处理
特定于范围的维护任务，例如拆分，合并和重新平衡

All Reads and writes are generally addressed to the replica holding
the lease; if none does, any replica may be addressed, causing it to try
to obtain the lease synchronously. Requests received by a non-lease holder
(for the HLC timestamp specified in the request's header) fail with an
error pointing at the replica's last known lease holder. These requests
are retried transparently with the updated lease by the gateway node and
never reach the client.

Since reads bypass Raft, a new lease holder will, among other things, ascertain
that its timestamp cache does not report timestamps smaller than the previous
lease holder's (so that it's compatible with reads which may have occurred on
the former lease holder). This is accomplished by letting leases enter
a <i>stasis period</i> (which is just the expiration minus the maximum clock
offset) before the actual expiration of the lease, so that all the next lease
holder has to do is set the low water mark of the timestamp cache to its
new lease's start time.

As a lease enters its stasis period, no more reads or writes are served, which
is undesirable. However, this would only happen in practice if a node became
unavailable. In almost all practical situations, no unavailability results
since leases are usually long-lived (and/or eagerly extended, which can avoid
the stasis period) or proactively transferred away from the lease holder, which
can also avoid the stasis period by promising not to serve any further reads
until the next lease goes into effect.

所有的读写操作通常都是针对副本控制
租约;如果没有，可以解决任何副本，导致它尝试
同步获得租约。非租赁持有人收到的请求
（对于请求标题中指定的HLC时间戳）将失败，并显示一个
错误指向副本的最后一个已知租约持有者。这些请求
透明地重试网关节点更新的租约，
永远不要到达客户端。

由于阅读绕过筏，新的租赁持有人，除其他事项外，将确定
它的时间戳缓存不会报告比以前更小的时间戳
租约持有人（以便它与可能发生的阅读相兼容）
前租约持有人）。这是通过让租赁进入
一个停滞期（这只是过期减去最大时钟
抵消）在租约实际到期之前，以便下一个租约
持有者需要做的是将时间戳缓存的低水位设置为其低位
新租约的开始时间。

当租约进入停滞期时，不再提供更多的读写内容
是不可取的。但是，这只会在实践中发生，如果一个节点成为
不可用。在几乎所有的实际情况下，都没有不可用的结果
因为租约通常是长期的（和/或急切地延长，这可以避免
停滞期）或主动转移离开租赁持有人，这是
也可以通过承诺不再服务任何进一步的阅读来避免停滞期
直到下一个租约生效。

## Colocation with Raft leadership

The range lease is completely separate from Raft leadership, and so without
further efforts, Raft leadership and the Range lease might not be held by the
same Replica. Since it's expensive to not have these two roles colocated (the
lease holder has to forward each proposal to the leader, adding costly RPC
round-trips), each lease renewal or transfer also attempts to colocate them.
In practice, that means that the mismatch is rare and self-corrects quickly.

范围租约与筏领导完全分开，所以没有
进一步的努力，筏领导和范围租约可能不会被持有
相同的副本。 由于没有这两个角色共存是很昂贵的（
租约持有者必须将每个提议转交给领导者，这增加了昂贵的RPC
往返），每次租赁续约或转让也会尝试将其合并。
在实践中，这意味着不匹配很少，并且很快就会自我纠正。

## Command Execution Flow

This subsection describes how a lease holder replica processes a
read/write command in more details. Each command specifies (1) a key
(or a range of keys) that the command accesses and (2) the ID of a
range which the key(s) belongs to. When receiving a command, a node
looks up a range by the specified Range ID and checks if the range is
still responsible for the supplied keys. If any of the keys do not
belong to the range, the node returns an error so that the client will
retry and send a request to a correct range.

When all the keys belong to the range, the node attempts to
process the command. If the command is an inconsistent read-only
command, it is processed immediately. If the command is a consistent
read or a write, the command is executed when both of the following
conditions hold:

本小节描述了租赁持有人副本如何处理a
读/写命令更详细。 每个命令指定（1）一个键
（或一系列键），以及（2）a的ID
密钥所属的范围。 当收到一个命令时，一个节点
通过指定的范围ID查找范围并检查范围是否
仍然对提供的密钥负责。 如果任何键没有
属于范围，节点返回错误，以便客户端
重试并发送请求到正确的范围。

当所有密钥都属于该范围时，该节点尝试执行
处理命令。 如果该命令是不一致的只读
命令，它会立即处理。 如果命令是一致的
读取或写入，当以下两者都执行该命令时
条件成立：

- The range replica has a range lease.
- There are no other running commands whose keys overlap with
the submitted command and cause read/write conflict.

When the first condition is not met, the replica attempts to acquire
a lease or returns an error so that the client will redirect the
command to the current lease holder. The second condition guarantees that
consistent read/write commands for a given key are sequentially
executed.

When the above two conditions are met, the lease holder replica processes the
command. Consistent reads are processed on the lease holder immediately.
Write commands are committed into the Raft log so that every replica
will execute the same commands. All commands produce deterministic
results so that the range replicas keep consistent states among them.

When a write command completes, all the replica updates their response
cache to ensure idempotency. When a read command completes, the lease holder
replica updates its timestamp cache to keep track of the latest read
for a given key.

There is a chance that a range lease gets expired while a command is
executed. Before executing a command, each replica checks if a replica
proposing the command has a still lease. When the lease has been
expired, the command will be rejected by the replica.

- 范围副本有一个范围租约。
- 没有其他键与键重叠的运行命令
提交的命令并导致读/写冲突。

当第一个条件不满足时，副本尝试获取
租约或返回错误，以便客户端重定向
命令给当前的租赁持有人。第二个条件保证
针对给定键的一致读/写命令是顺序的
执行。

当满足上述两个条件时，租赁持有人副本处理
命令。一致的读数立即在租赁持有人处理。
写命令被提交到Raft日志中，以便每个副本
将执行相同的命令。所有命令都会产生确定性
结果，以便范围副本在它们之间保持一致的状态。

写入命令完成后，所有副本都会更新其响应
缓存以确保幂等性。当读取命令完成时，租约持有者
副本更新其时间戳缓存以跟踪最新的读取
对于给定的密钥。

在命令执行期间，范围租约有可能过期
执行。在执行命令之前，每个副本检查一个副本
提出命令仍有租约。当租约一直
过期，命令将被副本拒绝。

# Splitting / Merging Ranges

Nodes split or merge ranges based on whether they exceed maximum or
minimum thresholds for capacity or load. Ranges exceeding maximums for
either capacity or load are split; ranges below minimums for *both*
capacity and load are merged.

Ranges maintain the same accounting statistics as accounting key
prefixes. These boil down to a time series of data points with minute
granularity. Everything from number of bytes to read/write queue sizes.
Arbitrary distillations of the accounting stats can be determined as the
basis for splitting / merging. Two sensible metrics for use with
split/merge are range size in bytes and IOps. A good metric for
rebalancing a replica from one node to another would be total read/write
queue wait times. These metrics are gossipped, with each range / node
passing along relevant metrics if they’re in the bottom or top of the
range it’s aware of.

A range finding itself exceeding either capacity or load threshold
splits. To this end, the range lease holder computes an appropriate split key
candidate and issues the split through Raft. In contrast to splitting,
merging requires a range to be below the minimum threshold for both
capacity *and* load. A range being merged chooses the smaller of the
ranges immediately preceding and succeeding it.

Splitting, merging, rebalancing and recovering all follow the same basic
algorithm for moving data between roach nodes. New target replicas are
created and added to the replica set of source range. Then each new
replica is brought up to date by either replaying the log in full or
copying a snapshot of the source replica data and then replaying the log
from the timestamp of the snapshot to catch up fully. Once the new
replicas are fully up to date, the range metadata is updated and old,
source replica(s) deleted if applicable.

根据节点是否超过最大值或节点来分割或合并范围
容量或负载的最低阈值。超过最大值的范围
容量或负载分开;范围低于* both *
容量和负载被合并。

范围保持与会计核算相同的会计统计
前缀。这些归结为一分钟的时间序列的数据点
粒度。从字节数到读取/写入队列大小的一切。
会计统计的任意蒸馏可以被确定为
拆分/合并的基础。两个明智的指标用于
分割/合并是以字节和IOps为单位的范围大小。一个很好的度量标准
将副本从一个节点重新平衡到另一个节点将是完全读取/写入
排队等待时间。这些度量标准会随着每个范围/节点而变化
如果他们处于底部或顶部，则传递相关指标
它意识到的范围。

范围发现本身超过容量或负载阈值
分裂。为此，范围租约持有人计算一个合适的拆分键
候选人并通过Raft发布分手。与分裂相反，
合并要求范围低于两者的最小阈值
容量*和*负载。被合并的范围选择较小的一个
范围紧接在它之前和之后。

拆分，合并，再平衡和恢复都遵循相同的基本原则
在roach节点之间移动数据的算法。新的目标副本是
创建并添加到源范围的副本集。然后每个新的
通过重新播放日志或者完整地复制副本
复制源副本数据的快照，然后重播日志
从快照的时间戳完全赶上。一旦新的
副本完全是最新的，范围元数据更新且旧，
如果适用，删除源副本。

**Coordinator** (lease holder replica)

```
if splitting
  SplitRange(split_key): splits happen locally on range replicas and
  only after being completed locally, are moved to new target replicas.
else if merging
  Choose new replicas on same servers as target range replicas;
  add to replica set.
else if rebalancing || recovering
  Choose new replica(s) on least loaded servers; add to replica set.
```

```
如果分裂
   SplitRange（split_key）：分割在范围副本和本地发生
   只有在本地完成之后，才会转移到新的目标副本。
否则如果合并
   在与目标范围副本相同的服务器上选择新副本;
   添加到副本集。
否则，如果重新平衡||恢复
   在最少加载的服务器上选择新副本; 添加到副本集。
```

**New Replica**

*Bring replica up to date:*

```
if all info can be read from replicated log
  copy replicated log
else
  snapshot source replica
  send successive ReadRange requests to source replica
  referencing snapshot

if merging
  combine ranges on all replicas
else if rebalancing || recovering
  remove old range replica(s)
```

Nodes split ranges when the total data in a range exceeds a
configurable maximum threshold. Similarly, ranges are merged when the
total data falls below a configurable minimum threshold.

**TBD: flesh this out**: Especially for merges (but also rebalancing) we have a
range disappearing from the local node; that range needs to disappear
gracefully, with a smooth handoff of operation to the new owner of its data.

Ranges are rebalanced if a node determines its load or capacity is one
of the worst in the cluster based on gossipped load stats. A node with
spare capacity is chosen in the same datacenter and a special-case split
is done which simply duplicates the data 1:1 and resets the range
configuration metadata.

```
如果所有信息都可以从复制日志中读取
  复制复制日志
其他
  快照源副本
  将连续的ReadRange请求发送到源副本
  引用快照

如果合并
  结合所有副本上的范围
否则，如果重新平衡||恢复
  删除旧范围副本（s）
```

当范围内的总数据超过a时，节点将分割范围
可配置的最大阈值。同样，当范围合并时
总数据低于可配置的最小阈值。

** TBD：肉体这一点**：特别是对于合并（但也是重新平衡），我们有一个
范围从本地节点消失;该范围需要消失
优雅地将操作顺利交给新的数据所有者。

如果节点确定其负载或容量为1，则范围将重新平衡
基于闲话负载统计的群集中最差的。带有的节点
备用容量选择在同一个数据中心和一个特殊情况下分割
完成它只是复制数据1：1并重置范围
配置元数据。

# Node Allocation (via Gossip)

New nodes must be allocated when a range is split. Instead of requiring
every node to know about the status of all or even a large number
of peer nodes --or-- alternatively requiring a specialized curator or
master with sufficiently global knowledge, we use a gossip protocol to
efficiently communicate only interesting information between all of the
nodes in the cluster. What’s interesting information? One example would
be whether a particular node has a lot of spare capacity. Each node,
when gossiping, compares each topic of gossip to its own state. If its
own state is somehow “more interesting” than the least interesting item
in the topic it’s seen recently, it includes its own state as part of
the next gossip session with a peer node. In this way, a node with
capacity sufficiently in excess of the mean quickly becomes discovered
by the entire cluster. To avoid piling onto outliers, nodes from the
high capacity set are selected at random for allocation.

分割范围时必须分配新节点。 而不是要求
每个节点都知道所有甚至是大数量的状态
对等节点 - 或者 - 或者需要专门的策展人或者
掌握足够的全球知识，我们使用八卦协议来
有效地沟通所有的有趣信息
集群中的节点。 什么是有趣的信息？ 一个例子会
是一个特定的节点是否有大量的闲置容量。 每个节点，
当闲聊时，将每个八卦话题与自己的状态进行比较。 如果它是
自己的国家在某种程度上比最不感兴趣的项目“更有趣”
在最近看到的主题中，它包含了自己的状态作为其一部分
与对等节点的下一个八卦会话。 这样，一个节点就可以了
足以超过平均值的能力很快就会被发现
整个集群。 为了避免堆积到异常值，从节点
随机选择高容量组进行分配。

The gossip protocol itself contains two primary components:

- **Peer Selection**: each node maintains up to N peers with which it
  regularly communicates. It selects peers with an eye towards
  maximizing fanout. A peer node which itself communicates with an
  array of otherwise unknown nodes will be selected over one which
  communicates with a set containing significant overlap. Each time
  gossip is initiated, each nodes’ set of peers is exchanged. Each
  node is then free to incorporate the other’s peers as it sees fit.
  To avoid any node suffering from excess incoming requests, a node
  may refuse to answer a gossip exchange. Each node is biased
  towards answering requests from nodes without significant overlap
  and refusing requests otherwise.

  Peers are efficiently selected using a heuristic as described in
  [Agarwal & Trachtenberg (2006)](https://drive.google.com/file/d/0B9GCVTp_FHJISmFRTThkOEZSM1U/edit?usp=sharing).

  **TBD**: how to avoid partitions? Need to work out a simulation of
  the protocol to tune the behavior and see empirically how well it
  works.

- **Gossip Selection**: what to communicate. Gossip is divided into
  topics. Load characteristics (capacity per disk, cpu load, and
  state [e.g. draining, ok, failure]) are used to drive node
  allocation. Range statistics (range read/write load, missing
  replicas, unavailable ranges) and network topology (inter-rack
  bandwidth/latency, inter-datacenter bandwidth/latency, subnet
  outages) are used for determining when to split ranges, when to
  recover replicas vs. wait for network connectivity, and for
  debugging / sysops. In all cases, a set of minimums and a set of
  maximums is propagated; each node applies its own view of the
  world to augment the values. Each minimum and maximum value is
  tagged with the reporting node and other accompanying contextual
  information. Each topic of gossip has its own protobuf to hold the
  structured data. The number of items of gossip in each topic is
  limited by a configurable bound.

  For efficiency, nodes assign each new item of gossip a sequence
  number and keep track of the highest sequence number each peer
  node has seen. Each round of gossip communicates only the delta
  containing new items.

  八卦协议本身包含两个主要组件：

- **对等选择**：每个节点最多保持N个与之对等的对等体
  定期沟通。它倾向于选择同行
  最大化扇出。一个对等节点本身与一个通信
  将选择其他未知节点的数组
  与包含显着重叠的集合通信。每一次
  八卦被启动，每个节点的对等组被交换。每
  节点随后可以根据自己的需要随意组合其他同龄人。
  为了避免任何节点遭受多余的传入请求，一个节点
  可能拒绝回答八卦交流。每个节点都有偏差
  用于回答来自节点的请求而没有显着的重叠
  否则拒绝请求。

  使用如下所述的启发式方法有效地选择对等项
  [Agarwal＆Trachtenberg（2006）]（https://drive.google.com/file/d/0B9GCVTp_FHJISmFRTThkOEZSM1U/edit?usp=sharing）。

  ** TBD **：如何避免分区？需要制定一个模拟
  协议来调整行为并从经验上看它有多好
  作品。

- **八卦选择**：沟通的内容。闲话被分成
  主题。负载特性（每个磁盘的容量，CPU负载和
  状态[例如排水，确定，故障]）用于驱动节点
  分配。范围统计（范围读取/写入加载，缺失
  副本，不可用范围）和网络拓扑（机架间
  带宽/延迟，数据中心间带宽/延迟，子网
  中断）用于确定何时分割范围，何时
  恢复副本与等待网络连接，以及
  调试/ sysops。在所有情况下，一组最小值和一组值
  最大值被传播;每个节点都应用其自己的视图
  世界来增加价值。每个最小值和最大值都是
  用报告节点和其他伴随的上下文标记
  信息。八卦的每一个话题都有自己的原型
  结构化数据。每个主题中八卦的项目数量是
  受限于可配置的界限。

  为了提高效率，节点为每个新闻项目分配一个序列
  并记录每个对等体的最高序列号
  节点已经看到。每一轮八卦只传达三角洲
  包含新项目。

# Node and Cluster Metrics

Every component of the system is responsible for exporting interesting
metrics about itself. These could be histograms, throughput counters, or
gauges.

These metrics are exported for external monitoring systems (such as Prometheus)
via a HTTP endpoint, but CockroachDB also implements an internal timeseries
database which is stored in the replicated key-value map.

Time series are stored at Store granularity and allow the admin dashboard
to efficiently gain visibility into a universe of information at the Cluster,
Node or Store level. A [periodic background process](RFCS/20160901_time_series_culling.md)
culls older timeseries data, downsampling and eventually discarding it.

系统的每个组件都负责导出有趣的内容
有关自己的指标。 这些可能是直方图，吞吐量计数器或
仪表。

这些指标被导出用于外部监控系统（如Prometheus）
通过HTTP端点，但CockroachDB也实现了内部时间序列
数据库存储在复制的键值映射中。

时间序列以商店粒度存储并允许管理员仪表板
为了有效地了解群集中的一系列信息，
节点或商店级别。 [定期后台流程]（RFCS / 20160901_time_series_culling.md）
选择较早的时间序列数据，降采样并最终丢弃它。

# Key-prefix Accounting and Zones

Arbitrarily fine-grained accounting is specified via
key prefixes. Key prefixes can overlap, as is necessary for capturing
hierarchical relationships. For illustrative purposes, let’s say keys
specifying rows in a set of databases have the following format:

`<db>:<table>:<primary-key>[:<secondary-key>]`

In this case, we might collect accounting with
key prefixes:

`db1`, `db1:user`, `db1:order`,

Accounting is kept for the entire map by default.

任意细粒度的会计通过指定
关键前缀。 关键字前缀可以重叠，这对捕获是必要的
层次关系。 出于说明的目的，我们说键
指定一组数据库中的行具有以下格式：

`<分贝>：<TABLE><主键>[：<次级键>]`

在这种情况下，我们可能会收集会计
关键字前缀：

`db1`，`db1：user`，`db1：order`，

默认情况下会为整个地图保留会计。

## Accounting
to keep accounting for a range defined by a key prefix, an entry is created in
the accounting system table. The format of accounting table keys is:

`\0acct<key-prefix>`

In practice, we assume each node is capable of caching the
entire accounting table as it is likely to be relatively small.

Accounting is kept for key prefix ranges with eventual consistency for
efficiency. There are two types of values which comprise accounting:
counts and occurrences, for lack of better terms. Counts describe
system state, such as the total number of bytes, rows,
etc. Occurrences include transient performance and load metrics. Both
types of accounting are captured as time series with minute
granularity. The length of time accounting metrics are kept is
configurable. Below are examples of each type of accounting value.

为了保持对由关键字前缀定义的范围的记帐，创建一个条目
会计系统表。 会计表格的格式是：

'\0acct<键前缀>`

在实践中，我们假设每个节点都能够缓存
整个会计表，因为它可能会相对较小。

会计保留关键字前缀范围，最终一致性为
效率。 有两种类型的值包括会计：
计数和事件，因为缺乏更好的条款。 计数描述
系统状态，如总字节数，行数，
等等。发生包括瞬态性能和负载指标。 都
会计类型以分钟形式记录为时间序列
粒度。 时间会计度量标准的长度是
可配置的。 以下是每种会计价值的例子。

**System State Counters/Performance**

- Count of items (e.g. rows)
- Total bytes
- Total key bytes
- Total value length
- Queued message count
- Queued message total bytes
- Count of values \< 16B
- Count of values \< 64B
- Count of values \< 256B
- Count of values \< 1K
- Count of values \< 4K
- Count of values \< 16K
- Count of values \< 64K
- Count of values \< 256K
- Count of values \< 1M
- Count of values \> 1M
- Total bytes of accounting


**Load Occurrences**

- Get op count
- Get total MB
- Put op count
- Put total MB
- Delete op count
- Delete total MB
- Delete range op count
- Delete range total MB
- Scan op count
- Scan op MB
- Split count
- Merge count

Because accounting information is kept as time series and over many
possible metrics of interest, the data can become numerous. Accounting
data are stored in the map near the key prefix described, in order to
distribute load (for both aggregation and storage).

Accounting keys for system state have the form:
`<key-prefix>|acctd<metric-name>*`. Notice the leading ‘pipe’
character. It’s meant to sort the root level account AFTER any other
system tables. They must increment the same underlying values as they
are permanent counts, and not transient activity. Logic at the
node takes care of snapshotting the value into an appropriately
suffixed (e.g. with timestamp hour) multi-value time series entry.

Keys for perf/load metrics:
`<key-prefix>acctd<metric-name><hourly-timestamp>`.

`<hourly-timestamp>`-suffixed accounting entries are multi-valued,
containing a varint64 entry for each minute with activity during the
specified hour.

To efficiently keep accounting over large key ranges, the task of
aggregation must be distributed. If activity occurs within the same
range as the key prefix for accounting, the updates are made as part
of the consensus write. If the ranges differ, then a message is sent
to the parent range to increment the accounting. If upon receiving the
message, the parent range also does not include the key prefix, it in
turn forwards it to its parent or left child in the balanced binary
tree which is maintained to describe the range hierarchy. This limits
the number of messages before an update is visible at the root to `2*log N`,
where `N` is the number of ranges in the key prefix.

因为会计信息以时间序列和许多方式保存
可能的利益指标，数据可能会变得很多。会计
数据存储在地图附近的密钥前缀描述中，以便于
分配负载（用于聚合和存储）。

系统状态的会计键具有以下形式：
`<键前缀> | acctd <度量标准名称> *`。注意领先的“管道”
字符。它意味着在任何其他类别之后对根级帐户进行排序
系统表。他们必须像他们一样增加相同的基础价值
是永久计数，而不是暂时活动。逻辑在
节点负责将值快照到适当的位置
后缀（例如时间戳小时）多值时间序列条目。

性能/负载指标的关键字：
`<键前缀> acctd <度量标准名称> <每小时时间戳>`。

`<hourly-timestamp>` - 后缀会计分录是多值的，
每分钟包含一个varint64条目，其中包含活动
指定小时。

为了有效地保持大范围的会计核算，任务
聚合必须是分布式的。如果活动发生在同一个内部
范围作为会计的关键前缀，更新将作为部分进行
的共识写。如果范围不同，则发送消息
到父范围以增加会计。如果收到的话
消息，父范围也不包含密钥前缀，它在
将其转交给平衡二元组中的父母或左孩子
树被维护来描述范围层次结构。这限制了
更新之前的消息数量在根目录可见为'2 * log N`，
其中`N`是键前缀中的范围数。

## Zones
zones are stored in the map with keys prefixed by
`\0zone` followed by the key prefix to which the zone
configuration applies. Zone values specify a protobuf containing
the datacenters from which replicas for ranges which fall under
the zone must be chosen.

Please see [pkg/config/config.proto](https://github.com/cockroachdb/cockroach/blob/master/pkg/config/config.proto) for up-to-date data structures used, the best entry point being `message ZoneConfig`.

If zones are modified in situ, each node verifies the
existing zones for its ranges against the zone configuration. If
it discovers differences, it reconfigures ranges in the same way
that it rebalances away from busy nodes, via special-case 1:1
split to a duplicate range comprising the new configuration.

区域以键为前缀的地图存储在地图中
`\ 0zone`后跟该区域的关键字前缀
配置适用。 区域值指定了一个包含protobuf的区域
数据中心从哪个范围复制到哪个数据中心
必须选择该区域。

请参阅[pkg / config / config.proto]（https://github.com/cockroachdb/cockroach/blob/master/pkg/config/config.proto）了解所用的最新数据结构，最佳入口点 作为'消息ZoneConfig`。

如果区域被原地修改，则每个节点都会验证该区域
现有区域的范围与区域配置相对应。 如果
它发现差异，它以相同的方式重新配置范围
它通过特殊情况1：1重新平衡远离繁忙节点
拆分成包含新配置的重复范围。

# SQL

Each node in a cluster can accept SQL client connections. CockroachDB
supports the PostgreSQL wire protocol, to enable reuse of native
PostgreSQL client drivers. Connections using SSL and authenticated
using client certificates are supported and even encouraged over
unencrypted (insecure) and password-based connections.

Each connection is associated with a SQL session which holds the
server-side state of the connection. Over the lifespan of a session
the client can send SQL to open/close transactions, issue statements
or queries or configure session parameters, much like with any other
SQL database.

群集中的每个节点都可以接受SQL客户端连接。CockroachDB
支持PostgreSQL有线协议，以便重用本机
PostgreSQL客户端驱动程序。 使用SSL和认证的连接
支持使用客户端证书，甚至鼓励
未加密（不安全）和基于密码的连接。

每个连接都与一个SQL会话关联，该会话持有该会话
服务器端的连接状态。 在会话的整个生命周期中
客户端可以发送SQL来打开/关闭事务，发出语句
或者查询或者配置会话参数，就像其他的一样
SQL数据库。

## Language support

CockroachDB also attempts to emulate the flavor of SQL supported by
PostgreSQL, although it also diverges in significant ways:

- CockroachDB exclusively implements MVCC-based consistency for
  transactions, and thus only supports SQL's isolation levels SNAPSHOT
  and SERIALIZABLE.  The other traditional SQL isolation levels are
  internally mapped to either SNAPSHOT or SERIALIZABLE.

- CockroachDB implements its own [SQL type system](RFCS/20160203_typing.md)
  which only supports a limited form of implicit coercions between
  types compared to PostgreSQL. The rationale is to keep the
  implementation simple and efficient, capitalizing on the observation
  that 1) most SQL code in clients is automatically generated with
  coherent typing already and 2) existing SQL code for other databases
  will need to be massaged for CockroachDB anyways.

  CockroachDB也尝试模拟SQL支持的风格
  PostgreSQL虽然在很大程度上也存在分歧：

  - CockroachDB专门为实现基于MVCC的一致性
     事务，因此只支持SQL的隔离级别SNAPSHOT
     和SERIALIZABLE。 其他传统的SQL隔离级别是
     内部映射到SNAPSHOT或SERIALIZABLE。

  - CockroachDB实现自己的[SQL类型系统]（RFCS / 20160203_typing.md）
     它只支持有限形式的隐式强制
     类型与PostgreSQL相比。 理由是保持
     实施简单高效，利用观察
     1）客户端中的大部分SQL代码都是自动生成的
     一致的打字已经和2）其他数据库的现有SQL代码
     无论如何都需要为CockroachDB进行按摩。

## SQL architecture

Client connections over the network are handled in each node by a
pgwire server process (goroutine). This handles the stream of incoming
commands and sends back responses including query/statement results.
The pgwire server also handles pgwire-level prepared statements,
binding prepared statements to arguments and looking up prepared
statements for execution.

Meanwhile the state of a SQL connection is maintained by a Session
object and a monolithic `planner` object (one per connection) which
coordinates execution between the session, the current SQL transaction
state and the underlying KV store.

Upon receiving a query/statement (either directly or via an execute
command for a previously prepared statement) the pgwire server forwards
the SQL text to the `planner` associated with the connection. The SQL
code is then transformed into a SQL query plan.
The query plan is implemented as a tree of objects which describe the
high-level data operations needed to resolve the query, for example
"join", "index join", "scan", "group", etc.

The query plan objects currently also embed the run-time state needed
for the execution of the query plan. Once the SQL query plan is ready,
methods on these objects then carry the execution out in the fashion
of "generators" in other programming languages: each node *starts* its
children nodes and from that point forward each child node serves as a
*generator* for a stream of result rows, which the parent node can
consume and transform incrementally and present to its own parent node
also as a generator.

The top-level planner consumes the data produced by the top node of
the query plan and returns it to the client via pgwire.

通过网络的客户端连接在每个节点中由a处理
pgwire服务器进程（goroutine）。这处理传入的流
命令并返回包括查询/语句结果的响应。
pgwire服务器还处理pgwire级准备好的语句，
将准备好的陈述与论据绑定并查找准备
执行语句。

同时，SQL连接的状态由Session维护
对象和一个单一的“规划器”对象（每个连接一个）
协调会话之间的执行，即当前的SQL事务
州和底层的KV商店。

一旦收到查询/陈述（直接或通过执行
命令为先前准备好的语句）pgwire服务器转发
将SQL文本添加到与连接关联的“计划者”中。 SQL
代码然后转换成SQL查询计划。
查询计划实现为描述该对象的对象树
例如，解析查询所需的高级数据操作
“加入”，“索引加入”，“扫描”，“组”等。

查询计划对象当前还嵌入了所需的运行时状态
用于执行查询计划。一旦SQL查询计划准备就绪，
然后这些对象的方法以时尚的方式执行
其他编程语言中的“生成器”：每个节点*启动*其
孩子节点，并从那个点向前，每个孩子节点作为一个
* generator *用于父节点可以的结果行流
消耗和增量转换并呈现​​给它自己的父节点
也作为发电机。

顶层规划器使用顶层节点生成的数据
查询计划并通过pgwire将其返回给客户端。

## Data mapping between the SQL model and KV

Every SQL table has a primary key in CockroachDB. (If a table is created
without one, an implicit primary key is provided automatically.)
The table identifier, followed by the value of the primary key for
each row, are encoded as the *prefix* of a key in the underlying KV
store.

Each remaining column or *column family* in the table is then encoded
as a value in the underlying KV store, and the column/family identifier
is appended as *suffix* to the KV key.

For example:

- after table `customers` is created in a database `mydb` with a
primary key column `name` and normal columns `address` and `URL`, the KV pairs
to store the schema would be:

每个SQL表在CockroachDB中都有一个主键。 （如果创建了一个表
没有一个，会自动提供一个隐含的主键。）
表标识符，后跟主键的值
每行都被编码为底层KV中的一个键的*前缀*
商店。

然后对表中剩余的每列或*列族*进行编码
作为基础KV商店中的价值，以及列/族标识符
作为*后缀*附加到KV键。

例如：

- 在表'客户`是在数据库`mydb`中创建的
主键列`name`和普通列`地址`和`URL`，KV对
存储架构将是：

| Key                          | Values |
| ---------------------------- | ------ |
| `/system/databases/mydb/id`  | 51     |
| `/system/tables/customer/id` | 42     |
| `/system/desc/51/42/address` | 69     |
| `/system/desc/51/42/url`     | 66     |

(The numeric values on the right are chosen arbitrarily for the
example; the structure of the schema keys on the left is simplified
for the example and subject to change.)  Each database/table/column
name is mapped to a spontaneously generated identifier, so as to
simplify renames.

Then for a single row in this table:

（右边的数字值是任意选择的
例; 左侧的模式键的结构被简化了
为示例并可能更改。）每个数据库/表/列
名称被映射到自发生成的标识符，以便于
简化重命名。

然后对于此表中的单个行：

| Key               | Values                           |
| ----------------- | -------------------------------- |
| `/51/42/Apple/69` | `1 Infinite Loop, Cupertino, CA` |
| `/51/42/Apple/66` | `http://apple.com/`              |

Each key has the table prefix `/51/42` followed by the primary key
prefix `/Apple` followed by the column/family suffix (`/66`,
`/69`). The KV value is directly encoded from the SQL value.

Efficient storage for the keys is guaranteed by the underlying RocksDB engine
by means of prefix compression.

Finally, for SQL indexes, the KV key is formed using the SQL value of the
indexed columns, and the KV value is the KV key prefix of the rest of
the indexed row.

每个键都有表前缀`/ 51 / 42`，后跟主键
前缀`/ Apple`后跟列/族后缀（`/ 66`，
`/69`）。 KV值是从SQL值直接编码的。

基础的RocksDB引擎可以保证密钥的高效存储
通过前缀压缩。

最后，对于SQL索引，KV键是使用SQL的SQL值形成的
索引列，而KV值是其余部分的KV关键字前缀
索引行。

## Distributed SQL

Dist-SQL is a new execution framework being developed as of Q3 2016 with the
goal of distributing the processing of SQL queries.
See the [Distributed SQL
RFC](RFCS/20160421_distributed_sql.md)
for a detailed design of the subsystem; this section will serve as a summary.

Distributing the processing is desirable for multiple reasons:
- Remote-side filtering: when querying for a set of rows that match a filtering
  expression, instead of querying all the keys in certain ranges and processing
  the filters after receiving the data on the gateway node over the network,
  we'd like the filtering expression to be processed by the lease holder or
  remote node, saving on network traffic and related processing.
- For statements like `UPDATE .. WHERE` and `DELETE .. WHERE` we want to
  perform the query and the updates on the node which has the data (as opposed
  to receiving results at the gateway over the network, and then performing the
  update or deletion there, which involves additional round-trips).
- Parallelize SQL computation: when significant computation is required, we
  want to distribute it to multiple node, so that it scales with the amount of
  data involved. This applies to `JOIN`s, aggregation, sorting.

The approach we took  was originally inspired by
[Sawzall](https://cloud.google.com/dataflow/model/programming-model) - a
project by Rob Pike et al. at Google that proposes a "shell" (high-level
language interpreter) to ease the exploitation of MapReduce. It provides a
clear separation between "local" processes which process a limited amount of
data and distributed computations, which are abstracted away behind a
restricted set of conceptual constructs.

To run SQL statements in a distributed fashion, we introduce a couple of concepts:
- _logical plan_ - similar on the surface to the `planNode` tree described in
  the [SQL](#sql) section, it represents the abstract (non-distributed) data flow
  through computation stages.
- _physical plan_ - a physical plan is conceptually a mapping of the _logical
  plan_ nodes to CockroachDB nodes. Logical plan nodes are replicated and
  specialized depending on the cluster topology. The components of the physical
  plan are scheduled and run on the cluster.

  Dist-SQL是2016年第三季度开发的一个新的执行框架
  分配SQL查询处理的目标。
  请参阅[分布式SQL
  RFC（RFCS / 20160421_distributed_sql.md）
  详细设计子系统;本节将作为总结。

  由于多种原因，分发处理是可取的：
  - 远程端过滤：查询与过滤相匹配的一组行时
    表达式，而不是查询特定范围和处理中的所有键
    在通过网络在网关节点上接收到数据之后的过滤器，
    我们希望过滤表达式由租赁持有者处理
    远程节点，节省网络流量和相关处理。
  - 对于像`UPDATE .. WHERE`和`DELETE .. WHERE`这样的语句，我们想要
    在具有数据的节点上执行查询和更新（相反，
    通过网络在网关处接收结果，然后执行
    更新或删除，其中涉及额外的往返）。
  - 并行SQL计算：当需要大量计算时，我们
    想要将它分发到多个节点，以便它与数量成比例
    涉及的数据。这适用于`JOIN`，聚合，排序。

  我们采取的方法最初是受到启发
  [Sawzall]（https://cloud.google.com/dataflow/model/programming-model） - a
  Rob Pike等人的项目。在谷歌提出了一个“壳”（高级别
  语言翻译）来缓解MapReduce的开发。它提供了一个
  明确区分处理有限数量的“本地”流程
  数据和分布式计算，它们被抽象出来
  有限的一组概念结构。

  为了以分布式方式运行SQL语句，我们引入了一些概念：
  - _logical plan_ - 类似于描述在`planNode`树上的表面
    [SQL]（＃sql）部分，它表示抽象（非分布式）数据流
    通过计算阶段。
  - _物理计划_ - 物理计划在概念上是_逻辑的映射
    plan_节点到CockroachDB节点。逻辑计划节点被复制和
    取决于集群拓扑。物理组件
    计划将在群集上进行计划和运行。

## Logical planning

The logical plan is made up of _aggregators_. Each _aggregator_ consumes an
_input stream_ of rows (or multiple streams for joins) and produces an _output
stream_ of rows. Both the input and the output streams have a set schema. The
streams are a logical concept and might not map to a single data stream in the
actual computation. Aggregators will be potentially distributed when converting
the *logical plan* to a *physical plan*; to express what distribution and
parallelization is allowed, an aggregator defines a _grouping_ on the data that
flows through it, expressing which rows need to be processed on the same node
(this mechanism constraints rows matching in a subset of columns to be
processed on the same node). This concept is useful for aggregators that need
to see some set of rows for producing output - e.g. the SQL aggregation
functions. An aggregator with no grouping is a special but important case in
which we are not aggregating multiple pieces of data, but we may be filtering,
transforming, or reordering individual pieces of data.

Special **table reader** aggregators with no inputs are used as data sources; a
table reader can be configured to output only certain columns, as needed.
A special **final** aggregator with no outputs is used for the results of the
query/statement.

To reflect the result ordering that a query has to produce, some aggregators
(`final`, `limit`) are configured with an **ordering requirement** on the input
stream (a list of columns with corresponding ascending/descending
requirements). Some aggregators (like `table readers`) can guarantee a certain
ordering on their output stream, called an **ordering guarantee**. All
aggregators have an associated **ordering characterization** function
`ord(input_order) -> output_order` that maps `input_order` (an ordering
guarantee on the input stream) into `output_order` (an ordering guarantee for
the output stream) - meaning that if the rows in the input stream are ordered
according to `input_order`, then the rows in the output stream will be ordered
according to `output_order`.

The ordering guarantee of the table readers along with the characterization
functions can be used to propagate ordering information across the logical plan.
When there is a mismatch (an aggregator has an ordering requirement that is not
matched by a guarantee), we insert a **sorting aggregator**.

逻辑计划由_aggregators_组成。每个_aggregator_消耗一个
_输入stream_行（或多个流的连接）并产生_output
行的流。输入流和输出流都有一个设置的模式。该
流是一个逻辑概念，可能不会映射到单个数据流中
实际计算。转换时可能会分发聚合器
*逻辑计划*到*物理计划*;表达什么分配和
并行化是允许的，聚合器在数据上定义一个_grouping_
流过它，表示哪些行需要在同一个节点上处理
（这个机制约束了行子集中的行
在同一个节点上处理）。这个概念对于需要的聚合器很有用
看一些产生输出的行 - 例如SQL聚合
功能。没有分组的聚合是一个特殊而重要的案例
我们不汇总多个数据片段，但我们可能会过滤，
转换或重新排序单个数据片段。

特殊**表格阅读器**没有输入的聚合器被用作数据源;一个
根据需要，表格阅读器可以配置为只输出某些列。
一个没有输出的特殊** final **聚合器用于结果
查询/语句。

为了反映查询必须产生的结果，一些聚合器
（`final`，`limit`）在输入上配置**排序要求**
流（具有相应升序/降序的列的列表
要求）。一些聚合器（如“表格阅读器”）可以保证一定的
在它们的输出流中排序，称为**排序保证**。所有
聚合器具有相关的**排序特征**功能
`ord（input_order） - > output_order`映射`input_order`（一个排序
保证输入流）到`output_order`（一个订单保证
输出流） - 意味着如果输入流中的行是有序的
根据`input_order`，输出流中的行将被排序
根据`output_order`。

表格阅读器的订购保证以及表征
函数可以用来在整个逻辑计划中传播订购信息。
当不匹配时（一个聚合器有一个不是的订货要求
匹配保证），我们插入**排序聚合器**。

### Types of aggregators

- `TABLE READER` is a special aggregator, with no input stream. It's configured
  with spans of a table or index and the schema that it needs to read.
  Like every other aggregator, it can be configured with a programmable output
  filter.
- `JOIN` performs a join on two streams, with equality constraints between
  certain columns. The aggregator is grouped on the columns that are
  constrained to be equal.
- `JOIN READER` performs point-lookups for rows with the keys indicated by the
  input stream. It can do so by performing (potentially remote) KV reads, or by
  setting up remote flows.
- `SET OPERATION` takes several inputs and performs set arithmetic on them
  (union, difference).
- `AGGREGATOR` is the one that does "aggregation" in the SQL sense. It groups
  rows and computes an aggregate for each group. The group is configured using
  the group key. `AGGREGATOR` can be configured with one or more aggregation
  functions:

  - TABLE READER是一个特殊的聚合器，没有输入流。 已配置
  具有表或索引的跨度以及它需要读取的模式。
  像其他所有聚合器一样，它可以配置一个可编程输出
  过滤。
- `JOIN`在两个流之间执行连接，其间具有相等的约束
  某些栏目。 聚合器分组在所在的列上
  被限制为相等。
- 'JOIN READER`用键指定的键对行进行点查找
  输入流。 它可以通过执行（可能是远程的）KV读取或通过
  设置远程流量。
- 'SET OPERATION'需要几个输入并对它们进行集算术运算
  （联盟，差异）。
- `AGGREGATOR`是SQL的意义上的“聚合”。 它分组
  行并计算每个组的聚合。 该组使用
  组密钥。 `AGGREGATOR`可以配置一个或多个聚合
  功能：

  - `SUM`
  - `COUNT`
  - `COUNT DISTINCT`
  - `DISTINCT`

  An optional output filter has access to the group key and all the
  aggregated values (i.e. it can use even values that are not ultimately
  outputted).
- `SORT` sorts the input according to a configurable set of columns.
  This is a no-grouping aggregator, hence it can be distributed arbitrarily to
  the data producers. This means that it doesn't produce a global ordering,
  instead it just guarantees an intra-stream ordering on each physical output
  streams). The global ordering, when needed, is achieved by an input
  synchronizer of a grouped processor (such as `LIMIT` or `FINAL`).
- `LIMIT` is a single-group aggregator that stops after reading so many input
  rows.
- `FINAL` is a single-group aggregator, scheduled on the gateway, that collects
  the results of the query. This aggregator will be hooked up to the pgwire
  connection to the client.

  一个可选的输出过滤器可以访问组密钥和所有
   聚合值（即它甚至可以使用最终不是的值）
  输出）。
- `SORT`按照一组可配置的列对输入进行排序。
   这是一个不分组的聚合器，因此它可以任意分配给
   数据生产者。 这意味着它不会产生全局排序，
   相反，它只是保证每个物理输出的流内排序
  流）。 全球排序，需要时，通过输入来实现
   分组处理器的同步器（如“LIMIT”或“FINAL”）。
- “LIMIT”是一个单组聚合器，它在阅读了如此多的输入后停止
  行。
- “FINAL”是收集在网关上的单组聚合器
   查询的结果。 这个聚合器将被连接到pgwire
   连接到客户端。

## Physical planning

Logical plans are transformed into physical plans in a *physical planning
phase*. See the [corresponding
section](RFCS/20160421_distributed_sql.md#from-logical-to-physical) of the Distributed SQL RFC
for details.  To summarize, each aggregator is planned as one or more
*processors*, which we distribute starting from the data layout - `TABLE
READER`s have multiple instances, split according to the ranges - each instance
is planned on the lease holder of the relevant range. From that point on,
subsequent processors are generally either colocated with their inputs, or
planned as singletons, usually on the final destination node.

逻辑计划转化为物理计划中的物理计划
相*。 看到[相应的
部分]（分布式SQL RFC的RFCS / 20160421_distributed_sql.md＃from-logical-to-physical）
了解详情。 总而言之，每个聚合器计划为一个或多个聚合器
*处理器*，我们从数据布局开始发布 - “TABLE”
READER`s有多个实例，根据范围分割 - 每个实例
计划在相关范围的租赁持有者身上。 从那时起，
后续的处理器通常或者与它们的输入共存，或者
计划为单身人士，通常在最终的目的地节点上。

### Processors

When turning a _logical plan_ into a _physical plan_, its nodes are turned into
_processors_. Processors are generally made up of three components:

![Processor](RFCS/images/distributed_sql_processor.png?raw=true "Processor")

1. The *input synchronizer* merges the input streams into a single stream of
   data. Types:
   * single-input (pass-through)
   * unsynchronized: passes rows from all input streams, arbitrarily
     interleaved.
   * ordered: the input physical streams have an ordering guarantee (namely the
     guarantee of the corresponding logical stream); the synchronizer is careful
     to interleave the streams so that the merged stream has the same guarantee.

2. The *data processor* core implements the data transformation or aggregation
   logic (and in some cases performs KV operations).

3. The *output router* splits the data processor's output to multiple streams;
   types:
   * single-output (pass-through)
   * mirror: every row is sent to all output streams
   * hashing: each row goes to a single output stream, chosen according
     to a hash function applied on certain elements of the data tuples.
   * by range: the router is configured with range information (relating to a
     certain table) and is able to send rows to the nodes that are lease holders for
     the respective ranges (useful for `JoinReader` nodes (taking index values
     to the node responsible for the PK) and `INSERT` (taking new rows to their
     lease holder-to-be)).

     将_logical plan_变成_physical plan_时，其节点将变成
     _processors_。处理器通常由三部分组成：

     ！[Processor]（RFCS / images / distributed_sql_processor.png？raw = true“Processor”）

     1. *输入同步器*将输入流合并为单个流
        数据。类型：
        *单输入（传递）
        *非同步：任意传递来自所有输入流的行
          交错。
        *订购：输入物理流有一个订购保证（即
          保证相应的逻辑流）;同步器很小心
          交错流，以便合并的流具有相同的保证。

     2. *数据处理器*内核实现数据转换或聚合
        逻辑（并且在某些情况下执行KV操作）。

     3. *输出路由器*将数据处理器的输出分成多个流;
        类型：
        *单路输出（直通）
        *镜像：每行都发送到所有输出流
        *散列：每行输出到一个输出流，根据选择
          到应用于数据元组的某些元素的散列函数。
        *按范围：路由器配置范围信息（与a有关
          某些表），并且能够将行发送到租赁持有者的节点
          各自的范围（对JoinReader节点有用（采用索引值）
          到负责PK的节点）和“INSERT”（将新行写入其中
          租赁持有人））。

To illustrate with an example from the Distributed SQL RFC, the query:



为了用来自分布式SQL RFC的示例进行说明，查询：
```
TABLE Orders (OId INT PRIMARY KEY, CId INT, Value DECIMAL, Date DATE)

SELECT CID, SUM(VALUE) FROM Orders
  WHERE DATE > 2015
  GROUP BY CID
  ORDER BY 1 - SUM(Value)
```

produces the following logical plan:

![Logical plan](RFCS/images/distributed_sql_logical_plan.png?raw=true "Logical Plan")

This logical plan above could be transformed into either one of the following
physical plans:

![Physical plan](RFCS/images/distributed_sql_physical_plan.png?raw=true "Physical Plan")

or

![Alternate physical plan](RFCS/images/distributed_sql_physical_plan_2.png?raw=true "Alternate physical Plan")

产生以下逻辑计划：

！[逻辑计划]（RFCS / images / distributed_sql_logical_plan.png？raw = true“逻辑计划”）

上述逻辑计划可以转换为以下任一种
实物计划：

！[物理计划]（RFCS / images / distributed_sql_physical_plan.png？raw = true“物理计划”）

要么

！[备用物理计划]（RFCS / images / distributed_sql_physical_plan_2.png？raw = true“备用物理计划”）


## Execution infrastructure

Once a physical plan has been generated, the system needs to divvy it up
between the nodes and send it around for execution. Each node is responsible
for locally scheduling data processors and input synchronizers. Nodes also
communicate with each other for connecting output routers to input
synchronizers through a streaming interface.

一旦生成了物理计划，系统需要将其分解
在节点之间传送并执行。 每个节点都有责任
用于本地调度数据处理器和输入同步器。 节点也
相互通信以将输出路由器连接到输入
同步器通过流媒体接口。

### Creating a local plan: the `ScheduleFlows` RPC

Distributed execution starts with the gateway making a request to every node
that's supposed to execute part of the plan asking the node to schedule the
sub-plan(s) it's responsible for (except for "on-the-fly" flows, see design
doc). A node might be responsible for multiple disparate pieces of the overall
DAG - let's call each of them a *flow*. A flow is described by the sequence of
physical plan nodes in it, the connections between them (input synchronizers,
output routers) plus identifiers for the input streams of the top node in the
plan and the output streams of the (possibly multiple) bottom nodes. A node
might be responsible for multiple heterogeneous flows. More commonly, when a
node is the lease holder for multiple ranges from the same table involved in
the query, it will run a `TableReader` configured with all the spans to be
read across all the ranges local to the node.

A node therefore implements a `ScheduleFlows` RPC which takes a set of flows,
sets up the input and output [mailboxes](#mailboxes), creates the local
processors and starts their execution.

分布式执行始于网关向每个节点发出请求
这应该执行计划的一部分，要求节点安排时间表
它负责的子计划（除了“即时”流程外，请参阅设计
DOC）。节点可能负责整体的多个不同部分
DAG - 让我们打电话给他们每个人*流*。流程由序列来描述
物理计划节点，它们之间的连接（输入同步器，
输出路由器），以及顶层节点输入流的标识符
计划和（可能多个）底层节点的输出流。节点
可能负责多种异构流程。更常见的是，当一个
节点是涉及的同一表中多个范围的租赁持有者
该查询将运行一个配置了所有跨度的“TableReader”
读取节点本地的所有范围。

因此节点实现了一个`ScheduleFlows` RPC，它接受一组流，
设置输入和输出[邮箱]（＃邮箱），创建本地
处理器并开始执行。

### Local scheduling of flows

The simplest way to schedule the different processors locally on a node is
concurrently: each data processor, synchronizer and router runs as a goroutine,
with channels between them. The channels are buffered to synchronize producers
and consumers to a controllable degree.

在节点上本地调度不同处理器的最简单方法是
同时：每个数据处理器，同步器和路由器作为goroutine运行，
与他们之间的渠道。 通道被缓冲以同步生产者
和消费者的可控程度。

### Mailboxes

Flows on different nodes communicate with each other over gRPC streams. To
allow the producer and the consumer to start at different times,
`ScheduleFlows` creates named mailboxes for all the input and output streams.
These message boxes will hold some number of tuples in an internal queue until
a gRPC stream is established for transporting them. From that moment on, gRPC
flow control is used to synchronize the producer and consumer. A gRPC stream is
established by the consumer using the `StreamMailbox` RPC, taking a mailbox id
(the same one that's been already used in the flows passed to `ScheduleFlows`).

A diagram of a simple query using mailboxes for its execution:
![Mailboxes](RFCS/images/distributed_sql_mailboxes.png?raw=true)


不同节点上的流量通过gRPC流彼此通信。 至
允许生产者和消费者在不同的时间开始，
`ScheduleFlows`为所有输入和输出流创建命名邮箱。
这些消息框将在内部队列中保存一定数量的元组
建立gRPC流来运输它们。 从那一刻起，gRPC
流量控制用于同步生产者和消费者。 gRPC流是
消费者使用`StreamMailbox` RPC建立一个邮箱ID
（与传递给`ScheduleFlows`的流中已经使用的相同）。

使用邮箱执行简单查询的图表：
！[邮箱]（RFCS/图像/ distributed_sql_mailboxes.png？原始=真）

## A complex example: Daily Promotion

To give a visual intuition of all the concepts presented, we draw the physical plan of a relatively involved query. The
point of the query is to help with a promotion that goes out daily, targeting
customers that have spent over $1000 in the last year. We'll insert into the
`DailyPromotion` table rows representing each such customer and the sum of her
recent orders.

为了给出所有提出的概念的视觉直觉，我们绘制了相对涉及的查询的物理计划。该
查询的要点是帮助每天进行的促销活动，定位
去年花费超过1000美元的顾客。 我们将插入到
代表每个此类客户的DailyPromotion表格行以及她的总和
最近的订单。

```SQL
TABLE DailyPromotion (
  Email TEXT,
  Name TEXT,
  OrderCount INT
)

TABLE Customers (
  CustomerID INT PRIMARY KEY,
  Email TEXT,
  Name TEXT
)

TABLE Orders (
  CustomerID INT,
  Date DATETIME,
  Value INT,

  PRIMARY KEY (CustomerID, Date),
  INDEX date (Date)
)

INSERT INTO DailyPromotion
(SELECT c.Email, c.Name, os.OrderCount FROM
      Customers AS c
    INNER JOIN
      (SELECT CustomerID, COUNT(*) as OrderCount FROM Orders
        WHERE Date >= '2015-01-01'
        GROUP BY CustomerID HAVING SUM(Value) >= 1000) AS os
    ON c.CustomerID = os.CustomerID)
```

A possible physical plan:
![Physical plan](RFCS/images/distributed_sql_daily_promotion_physical_plan.png?raw=true)
