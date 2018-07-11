
* oceanbase 支付宝

https://www.oschina.net/p/oceanbase
https://www.jianshu.com/p/fecf188733f6
https://github.com/alibaba/oceanbase
https://www.zhihu.com/question/19841579

* Volcano-An Extensible and Parallel Query Evaluation System
tidb executor 基于这个实现，postgre、spark sql 1都是基于这个实现
注重扩展性、并行性
每个算子都应该实现成一个iterator，有三个接口，分别是open/next/close，其中open可以做一些资源初始化，比如打开文件，next将游标向前推进返回Next-Record（由ID和指向buffer的数据指针组成；为了性能，可以返回一批记录），close负责销毁资源

https://zhuanlan.zhihu.com/p/3422091
https://www.iteblog.com/archives/1679.html

经典论文翻译导读之《Large-scale Incremental Processing Using Distributed Transactions and Notifications》
http://www.importnew.com/2896.html

* MySQL锁总结

https://zhuanlan.zhihu.com/p/29150809


分布式领域论文译序
https://www.cnblogs.com/superf0sh/p/5754283.html?utm_source=itdadao&utm_medium=referral
