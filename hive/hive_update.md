

[Hive:ORC File Format存储格式详解](https://www.iteblog.com/archives/1014.html)

[hive update 如何配置](https://blog.csdn.net/weixin_43215250/article/details/86151089)

[hive not designed for oltp and does not offer real-time queries](https://stackoverflow.com/questions/17810537/how-to-delete-and-update-a-record-in-hive)

https://issues.apache.org/jira/browse/HIVE-5317

Regarding use cases, it appears that this design won't be able to have fast performance for fine-grained inserts. ...
 -- Agreed, this will fail badly in a one insert at a time situation. That isn't what we're going after. We would like to be able to handle a batch inserts every minute, but for the moment that seems like the floor.
--> 那我们从binlog按行读取数据写入hive，性能是不行的?

But we can't do 500K update statements an hour, so it doesn't seem the ACID does us any good for this use case until we have merge

[hive transactions](https://cwiki.apache.org/confluence/display/Hive/Hive+Transactions)

Reading/writing to an ACID table from a non-ACID session is not allowed. In other words, the Hive transaction manager must be set to org.apache.hadoop.hive.ql.lockmgr.DbTxnManager in order to work with ACID tables.

## 实际测试

### 测试环境

* 硬件： 4G 1Core 虚拟机(nvme ssd) ×3 
* hadoop 2.7.7
* hive 2.1.1

```
$ hadoop version
Hadoop 2.7.7
Subversion Unknown -r c1aad84bd27cd79c3d1a7dd58202a8c3ee1ed3ac
Compiled by stevel on 2018-07-18T22:47Z
Compiled with protoc 2.5.0
From source with checksum 792e15d20b12c74bd6f19a1fb886490
This command was run using /data/deploy/hadoop-2.7.7/share/hadoop/common/hadoop-common-2.7.7.jar


$ hive --version
Hive 2.1.1
Subversion git://home/data/code/apache/hive -r 1af77bbf8356e86cabbed92cfa8cc2e1470a1d5c
Compiled by z on Wed Feb 5 09:17:50 CST 2020
From source with checksum 81e843b0ebe6d59f802ca641205579ba
```


#### 数据准备
```
create table if not exists uu(
  id bigint,
  sex tinyint  COMMENT '性别',
  name String COMMENT '姓名'
) COMMENT '用户表'
partitioned by (year string)
clustered by (id) into 8 buckets
row format delimited fields terminated by '\t'
stored as orc TBLPROPERTIES('transactional'='true');
```

```sql
insert into uu partition(year='2018') values (1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c');
insert into uu partition(year='2019') values (1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c');
insert into uu partition(year='2020') values (1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c'),(1,2,'a'),(2,3,'b'),(3,4,'c');

insert into uu partition(year='2017') (id,sex,name) select id,sex,name from uu;
insert into uu partition(year='2018') (id,sex,name) select id,sex,name from uu;
insert into uu partition(year='2019') (id,sex,name) select id,sex,name from uu;
insert into uu partition(year='2020') (id,sex,name) select id,sex,name from uu;
```

#### 测试结果

```
jdbc:hive2://master:10000> select count(1) from uu;
+--------+--+
|   c0   |
+--------+--+
| 11880 |
+--------+--+

update uu set name ='sdfs' ;
No rows affected (23.636 seconds)

0: jdbc:hive2://master:10000> update uu set name ='sdfs' ;
No rows affected (24.851 seconds)

```

```
jdbc:hive2://master:10000> select count(1) from uu;
+--------+--+
|   c0   |
+--------+--+
| 23760  |
+--------+--+

jdbc:hive2://master:10000> update uu set name ='sdfs' ;
No rows affected (24.328 seconds)

select count(1) from uu where year='2020';
+--------+--+
|   c0   |
+--------+--+
| 16776  |
+--------+--+
1 row selected (18.142 seconds)


update uu set name ='sdfs' where year='2020';
No rows affected (18.327 seconds)

update uu set name ='sdfs' where year='2020';
No rows affected (18.39 seconds)

```

```

select count(1) from uu ;
+---------+--+
|   c0    |
+---------+--+
| 190080  |
+---------+--+
1 row selected (45.773 seconds)

update uu set name='lixi';
No rows affected (26.685 seconds)

update uu set name='lixi';
No rows affected (41.091 seconds)

update uu set name='lixi';
No rows affected (28.064 seconds)

select count(1) from uu where year='2020';
+---------+--+
|   c0    |
+---------+--+
| 183096  |
+---------+--+
1 row selected (18.205 seconds)

update uu set name='lixi' where year='2020';
No rows affected (20.756 seconds)

update uu set name='lixi' where year='2020';
No rows affected (39.564 seconds)

update uu set name='lixi' where year='2020';
No rows affected (20.584 seconds)

update uu set name='lixi' where year='2020';
No rows affected (20.501 seconds)

update uu set name='lixi' where year='2020';
No rows affected (25.805 seconds)

update uu set name='lixi' where year='2020';
No rows affected (24.67 seconds)

```


update 60100条记录 100 次的时间(java程序jdbc做的，时间会比beeline的长)：
39,80,42,58,42,59,57,84,40,44,45,41,83,45,44,46,44,46,87,58,59,60,47,92,59,45,59,45,84,44,43,44,42,44,97,43,44,44,60,84,43,57,43,58,97,44,58,45,46,81,43,43,59,43,42,84,60,58,42,43,82,44,43,45,40,44,84,43,57,43,60,97,43,59,50,43,59,99,50,48,48,98,41,43,43,44,44,97,58,45,46,46,96,59,43,43,43,58,96,43

机械盘环境，之前是ssd的，可能速度差别主要在这方面
update 15万
114,111,125,105

update 118万
156,153,123

update 1000万
558,140,176,153,106
