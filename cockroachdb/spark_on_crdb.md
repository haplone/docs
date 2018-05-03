

use pg sql
```seq
client->spark: spark sql
spark->spark: spark sql to pg sql
spark-> cr(any): pg sql
cr(any)->cr node1: distsql
cr node1-->cr(any): table data
cr(any)->cr node2: distsql
cr node2-->cr(any): table data
cr(any)--> spark: table data
spark->spark: rdd compute
spark-->client: result
```


------------------------------------

use kv
```seq
spark AM-> cr(any):  metadata(grpc)
cr(any)--> spark AM: table,range,node
spark AM->spark AM: analysis sql
spark AM->yarn rm: request containters
yarn rm-->spark AM: containters by range location
spark AM->spark e1: request
spark e1->cr n1: start key,end key(grpc)
cr n1--> spark e1: kv
spark e1->spark e1: rdd compute
spark e1-->spark AM: rdd
spark AM->spark e_n: request
spark e_n->cr n_n: start key,end key(grpc)
cr n_n--> spark e_n: kv
spark e_n->spark e_n: rdd compute
spark e_n-->spark AM: rdd
```

use distSql (convert spark sql to pg sql for each range)
------------------------------------

```seq
client->spark: spark sql
spark-> cr(any): get metadata by grpc
cr(any)--> spark: return table,range,node
spark->spark: spark sql to pg sql for each range
spark->cr node1: get data by distsql
cr node1-->spark: table data
spark->cr node2: get data by distsql
cr node2-->spark: table data
spark->spark: rdd compute
spark-->client: result
```
