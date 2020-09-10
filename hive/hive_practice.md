

设置引擎

set hive.execution.engine=mr;  
set hive.execution.engine=spark;  
set hive.execution.engine=tez;  

如果使用的是mr(原生mapreduce)
SET mapreduce.job.queuename=etl;

如果使用的引擎是tez
set tez.queue.name=etl
设置队列（etl为队列名称，默认为default）


原文链接：https://blog.csdn.net/qbs946/article/details/80909046
