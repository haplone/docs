# kudu_storage_for_fast_analystics_on_fast_data

## Abstract

Kudu is an open source storage engine for `structured data` which supports `low-latency random access` together with efficient `analytical access` patterns. Kudu distributes data using `horizontal partitioning` and replicates each partition using `Raft consensus`, providing low mean-time-to-recovery and low tail latencies. Kudu is designed within the context of the Hadoop ecosystem and supports many modes of access via tools such as Cloudera Impala[20], Apache Spark[28], and MapReduce[17].

## 1 Introduction

In recent years, explosive growth in the amount of data being generated and captured by enterprises has resulted in the rapid adoption of open source technology which is able to store massive data sets at scale and at low cost. In particular, the Hadoop ecosystem has become a focal point for such “big data” workloads, because many traditional open source database systems have lagged in offering a scalable alternative.

Structured storage in the Hadoop ecosystem has typically been achieved in two ways: for static data sets, data is typically stored on HDFS using binary data formats such as Apache Avro[1] or Apache Parquet[3]. However, neither HDFS nor these formats has any provision for `updating individual records`, or for `efficient random access`. Mutable data sets are typically stored in semi-structured stores such as Apache HBase[2] or Apache Cassandra[21]. These systems allow for low-latency record-level reads and writes, but lag far behind the static file formats in terms of `sequential read throughput` for applications such as SQL-based analytics or machine learning.

The gap between the analytic performance offered by static data sets on HDFS and the low-latency row-level random access capabilities of HBase and Cassandra has required practitioners to develop complex architectures when the need for both access patterns arises in a single application. In particular, many of Cloudera’s customers have developed data pipelines which involve streaming ingest and updates in  HBase, followed by periodic jobs to export tables to Parquet for later analysis. Such architectures suffer several downsides:

1. Application architects must write complex code to manage the flow and synchronization of data between the two systems.

2. Operators must manage consistent backups, security policies, and monitoring across multiple distinct systems.

3. The resulting architecture may exhibit significant lag between the arrival of new data into the HBase “staging area” and the time when the new data is available for analytics.

4. In the real world, systems often need to accomodate late-arriving data, corrections on past records, or privacy-related deletions on data that has already been migrated to the immutable store. Achieving this may involve expensive rewriting and swapping of partitions and manual intervention.

Kudu is a new storage system designed and implemented from the ground up to fill this gap between high-throughput sequential-access storage systems such as HDFS[27] and low-latency random-access systems such as HBase or Cassandra. While these existing systems continue to hold advantages in some situations, Kudu offers a “happy medium” alternative that can dramatically simplify the architecture of many common workloads. In particular, Kudu offers a simple API for row-level inserts, updates, and deletes, while providing table scans at throughputs similar to Parquet, a commonly-used columnar format for static data.

This paper introduces the architecture of Kudu. Section 2 describes the system from a user’s point of view, introducing the data model, APIs, and operator-visible constructs. Section 3 describes the architecture of Kudu, including how it partitions and replicates data across nodes, recovers from faults, and performs common operations. Section 4 explains how Kudu stores its data on disk in order to combine fast random access with efficient analytics. Section 5 discusses integrations between Kudu and other Hadoop ecosystem projects. Section 6 presents preliminary performance results in synthetic workloads.

## 2 Kudu at a high level

### 2.1 Tables and schemas

From the perspective of a user, Kudu is a storage system for `tables of structured data`. A Kudu cluster may have any number of tables, each of which has a well-defined schema consisting of a finite number of columns. Each such column has a name, type (e.g INT32 or STRING) and optional nullability. Some ordered subset of those columns are specified to be the table’s primary key. The primary key enforces a uniqueness constraint (at most one row may have a given primary key tuple) and acts as the sole index by which rows may be efficiently updated or deleted. This data model is familiar to users of relational databases, but differs from many other distributed datastores such as Cassandra. MongoDB[6], Riak[8], BigTable[12], etc.

As with a relational database, the user must define the schema of a table at the time of creation. Attempts to insert data into undefined columns result in errors, as do violations of the primary key uniqueness constraint. The user may at any time issue an alter table command to add or drop columns, with the restriction that primary key columns cannot be dropped.

Our decision to explicitly specify types for columns instead of using a NoSQL-style “everything is bytes” is motivated by two factors:

1. Explicit types allow us to use type-specific columnar encodings such as bit-packing for integers.

2. Explicit types allow us to expose SQL-like metadata to other systems such as commonly used business intelligence or data exploration tools.

Unlike most relational databases, Kudu does not currently offer secondary indexes or uniqueness constraints other than the primary key. Currently, Kudu requires that every table has a primary key defined, though we anticipate that a future version will add automatic generation of surrogate keys.

### 2.2 Write operations

After creating a table, the user mutates the table using Insert, Update, and Delete APIs. In all cases, the user must fully specify a primary key – predicate-based deletions or updates must be handled by a higher-level access mechanism (see section 5).

Kudu offers APIs in Java and C++, with experimental support for Python. The APIs allow precise control over batching and asynchronous error handling to amortize the cost of round trips when performing bulk data operations (such as data loads or large updates). Currently, Kudu `does not offer any multi-row transactional APIs`: each mutation conceptually executes as its own transaction, despite being automatically batched with other mutations for better performance.Modifications within a single row are always executed atomically across columns.

### 2.3 Read operations

Kudu offers only a Scan operation to retrieve data from a table. On a scan, the user may add any number of predicates to
filter the results. Currently, we offer only two types of pred-
icates: comparisons between a column and a constant value,
and composite primary key ranges. These predicates are in-
terpreted both by the client API and the server to efficiently
cull the amount of data transferred from the disk and over
the network.

In addition to applying predicates, the user may specify a
projection for a scan. A projection consists of a subset of
columns to be retrieved. Because Kudu’s on-disk storage is
columnar, specifying such a subset can substantially improve
performance for typical analytic workloads.




