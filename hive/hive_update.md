

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

