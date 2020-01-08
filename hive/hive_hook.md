

hive 中已经实现了一些hook可以跟踪sql执行情况，特别是LineageLog已经实现得不错，能做到字段血缘程度。

## hook

主要是semantic-analyzer,driver-run,execution 三个阶段

Pre-semantic-analyzer hooks：在Hive在查询字符串上运行语义分析器之前调用。

Post-semantic-analyzer hooks：在Hive在查询字符串上运行语义分析器之后调用。

Pre-driver-run hooks：在driver执行查询之前调用。

Post-driver-run hooks：在driver执行查询之后调用。

Pre-execution hooks：在执行引擎执行查询之前调用。请注意，这个目的是此时已经为Hive准备了一个优化的查询计划。

Post-execution hooks：在查询执行完成之后以及将结果返回给用户之前调用。

Failure-execution hooks：当查询执行失败时调用。

### config

将LineageLogger添加到配置中
```shell
vim /usr/local/hive/conf/hive-site.xml
<property>
    <name>hive.exec.post.hooks</name>
    <value>org.apache.hadoop.hive.ql.hooks.LineageLogger</value>
</property>
```

配置hook输出
```shell
vim ${HIVE_HOME}/conf/hive-log4j2.properties
og4j.logger.org.apache.hadoop.hive.ql.hooks.LineageLogger=INFOo
```

可在hive.log 看结果


## 参考

[HIVE 字段级血缘分析 写入Neo4j](https://blog.csdn.net/xw514124202/article/details/94029564)

[利用LineageLogger分析HiveQL中的字段级别血缘关系 ](http://cxy7.com/articles/2017/11/10/1510310104765.html)

[利用LineageInfo分析HiveQL中的表级别血缘关系 ](http://cxy7.com/articles/2017/11/10/1510306038754.html)

基于LineageLogger进行扩展，支持更多的sql解析，我们上线前，需要确认我们的sql
[hive血缘关系之输入表与目标表的解析](https://www.cnblogs.com/wuxilc/p/9326130.html)

[Hive SQL运行状态监控（HiveSQLMonitor）](https://www.cnblogs.com/yurunmiao/p/4224137.html)

[你想了解的Hive Query生命周期--钩子函数篇](https://my.oschina.net/kavn/blog/1514648)
