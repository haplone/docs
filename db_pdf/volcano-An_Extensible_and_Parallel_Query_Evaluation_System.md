# Abstract-

To investigate the interactions of extensibility and
parallelism in database query processing, 
we have developed a new dataflow query execution system called Volcano.
The Volcano effort provides a rich environment for research and education 
in database systems design, heuristics for query optimization, 
parallel query execution, and resource allocation.


为了研究数据库查询处理可扩展性和并行的相互作用，
我们开发了一个新的数据流查询执行系统名为Volcano。
Volcano努力提供了丰富的环境用于数据库系统设计的研究和教育，
用于查询优化，并行查询执行和资源分配的启发式方法。


Volcano uses a standard interface between algebra operators,
 allowing easy addition of new operators and operator implementations.
Operations on individual items, e.g., predicates,
 are imported into the query processing operators using support functions.
The semantics of support functions is not prescribed; 
any data type including complex objects and any operation can be realized. 
Thus, Volcano is extensible with new operators, algorithms,
data types, and type-specific methods.


Volcano使用代数运算符之间的标准接口，允许轻松添加新的操作和操作实现。
对单个项目的操作，例如谓词，使用支持函数导入到查询处理运算符中。
没有规定支持函数的语义;可以实现包括复杂对象和任何操作的任何数据类型。
因此，Volcano可以通过新的算子，算法，数据类型和特定于类型的方法进行扩展。


Volcano includes two novel meta-operators.
The choose-plan meta-operator supports dynamic query evaluation plans 
that allow delaying selected optimization decisions until run-time,
e.g., for embedded queries with free variables. 
The exchange meta-operator supports intra-operator parallelism
on partitioned datasets and both vertical and horizontal inter-operator
parallelism, translating between demand-driven dataflow 
within processes and data-driven dataflow between processes.


Volcano包括两个新的元运算符。
choose-plan元运算符支持动态查询评估计划
允许延迟选定的优化决策直到运行时，例如，对于具有自由变量的嵌入式查询。
exchange 元运算符支持在分区数据集以及垂直和水平内部运算符并行，支持intra-opertor并行。
在需求驱动数据流操作和数据驱动数据流操作间进行转换。

All operators, with the exception of the exchange operator,
have been designed and implemented in a single-process environment, 
and parallelized using the exchange operator. 
Even operators not yet designed can be parallelized using this new
operator if they use and provide the interator interface. 
Thus, the issues of data manipulation and parallelism
have become orthogonal, making Volcano the first implemented
query execution engine that effectively combines extensibility and parallelism. 


所有操作中，除了exchange operator，都是设计并实现在一个单线程环境中，
然后通过exchange operator并行。
如果使用iterator接口，即使operator还设计好也可以通过这个operator并行化。
这样数据处理和并行化，就可以正交。这样使得Volcano成为第一个有效结合可扩展性和并行性的查询执行引擎。


