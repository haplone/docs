
# oceanbase 支付宝

https://www.oschina.net/p/oceanbase
https://www.jianshu.com/p/fecf188733f6
https://github.com/alibaba/oceanbase
https://www.zhihu.com/question/19841579

# Volcano-An Extensible and Parallel Query Evaluation System

tidb executor 基于这个实现，postgre、spark sql 1都是基于这个实现
注重扩展性、并行性
每个算子都应该实现成一个iterator，有三个接口，分别是open/next/close，其中open可以做一些资源初始化，比如打开文件，next将游标向前推进返回Next-Record（由ID和指向buffer的数据指针组成；为了性能，可以返回一批记录），close负责销毁资源

https://zhuanlan.zhihu.com/p/3422091
https://www.iteblog.com/archives/1679.html

经典论文翻译导读之《Large-scale Incremental Processing Using Distributed Transactions and Notifications》
http://www.importnew.com/2896.html

# MySQL锁总结

https://zhuanlan.zhihu.com/p/29150809


# join algorithms

* nested loop join
* block nested loop join
* merge join (sort-merge join)
* hash join
* index join

|类别|	Nested Loop|	Hash Join|	Merge Join|
|----|----|----|----|
|使用条件	| 任何条件 |	等值连接（=）|	等值或非等值连接(>，<，=，>=，<=)，‘<>’除外|
|相关资源|	CPU、磁盘I/O|	内存、临时空间	|内存、临时空间|
|特点|当有高选择性索引或进行限制性搜索时效率比较高，能够快速返回第一次的搜索结果。|当缺乏索引或者索引条件模糊时，Hash Join比Nested Loop有效。通常比Merge Join快。在数据仓库环境下，如果表的纪录数多，效率高。| 当缺乏索引或者索引条件模糊时，Merge Join比Nested Loop有效。非等值连接时，Merge Join比Hash Join更有效|
|缺点|当索引丢失或者查询条件限制不够时，效率很低；当表的纪录数多时，效率低。|为建立哈希表，需要大量内存。第一次的结果返回较慢。|所有的表都需要排序。它为最优化的吞吐量而设计，并且在结果没有全部找到前不返回数据。|

## nested loop join

适用于outer table(有的地方叫Master table)的记录集比较少(<10000)而且inner table(有的地方叫Detail table)索引选择性较好的情况下(inner table要有index)。

Nested Loops常执行Inner Join(内部联接)、Left Outer Join(左外部联接)、Left Semi Join(左半部联接)和Left Anti Semi Join(左反半部联接)逻辑操作。

## merge join

用在数据没有索引但是已经排序的情况下。

Merge Join第一个步骤是确保两个关联表都是按照关联的字段进行排序。如果关联字段有可用的索引，并且排序一致，则可以直接进行Merge Join操作


Merge Join常执行Inner Join(内部联接)、Left Outer Join(左外部联接)、Left Semi Join(左半部联接)、Left Anti Semi Join(左反半部联接)、Right Outer Join(右外部联接)、Right Semi Join(右半部联接)、Right Anti Semi Join(右反半部联接)和Union(联合)逻辑操作。

## hash join

Hash Match有两个输入：build input（也叫做outer input）和probe input（也叫做inner input），不仅用于inner/left/right join等，象union/group by等也会使用hash join进行操作，在group by中build input和probe input都是同一个记录集。


Hash Match操作分两个阶段完成：Build（构造）阶段和Probe（探测）阶段。

Build（构造）阶段主要构造哈希表(hash table)。在inner/left/right join等操作中，表的关联字段作为hash key；在group by操作中，group by的字段作为hash key；在union或其它一些去除重复记录的操作中，hash key包括所有的select字段。

Build操作从build input输入中取出每一行记录，将该行记录关联字段的值使用hash函数生成hash值，这个hash值对应到hash table中的hash buckets（哈希表目）。如果一个hash值对应到多个hash buckts，则这些hash buckets使用链表数据结构连接起来。当整个build input的table处理完毕后，build input中的所有记录都被hash table中的hash buckets引用/关联了。

Probe（探测）阶段，SQL Server从probe input输入中取出每一行记录，同样将该行记录关联字段的值，使用build阶段中相同的hash函数生成hash值，根据这个hash值，从build阶段构造的hash table中搜索对应的hash bucket。hash算法中为了解决冲突，hash bucket可能会链接到其它的hash bucket，probe动作会搜索整个冲突链上的hash bucket，以查找匹配的记录。


适用于两个表的数据量差别很大。但需要注意的是：如果HASH表太大，无法一次构造在内存中，则分成若干个partition，写入磁盘的temporary segment，则会多一个I/O的代价，会降低效率，此时需要有较大的temporary segment从而尽量提高I/O的性能。

[TiDB 源码阅读系列文章（九）Hash Join](https://zhuanlan.zhihu.com/p/37773956)

[TiDB 源码阅读系列文章（十一）Index Lookup Join](https://zhuanlan.zhihu.com/p/38572730)

[数据库join实现](https://blog.csdn.net/wankunde/article/details/78203080)

[SQL优化（一） Merge Join vs. Hash Join vs. Nested Loop](http://www.jasongj.com/2015/03/07/Join1/)


分布式领域论文译序
https://www.cnblogs.com/superf0sh/p/5754283.html?utm_source=itdadao&utm_medium=referral
