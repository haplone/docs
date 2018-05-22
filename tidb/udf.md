# tidb2.0 内置函数编写

在tidb2.0 如果我们想要实现内置函数,其实还是比较方便的。简单三步就行：
* 内置函数名称常量定义
* 根据接口实现业务逻辑（XXXXFunction,builtinXXXXSig)
* 注册内置函数


## 定义udf名称常量

在ast/functions.go中定义常量，如：
```go
const (
	// ...
	Length          = "length"
	// 这个就是我们新定义的
	UdfLength		= "udf_length"
)
```

## 在expression包中定义具体方法

* 定义XXXXFunction，使用getFunction提供udf具体实现结构体的初始化
* 定义builtinXXXXSig，实现evalXXX（XXX为数据返回类型）,实现业务逻辑

```go
// expression/builtin_string.go

var (
	_ functionClass = &lengthFunctionClass{}
	_ functionClass = &udfLengthFunctionClass{}
	// ...
)

var (
	_ builtinFunc = &builtinLengthSig{}
	_ builtinFunc = &builtinUdfLengthSig{}
	// ...
)

// 之前的实现
type lengthFunctionClass struct {
	baseFunctionClass
}

func (c *lengthFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, errors.Trace(err)
	}
	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETString)
	bf.tp.Flen = 10
	sig := &builtinLengthSig{bf}
	return sig, nil
}

type builtinLengthSig struct {
	baseBuiltinFunc
}

func (b *builtinLengthSig) Clone() builtinFunc {
	newSig := &builtinLengthSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

// evalInt evaluates a builtinLengthSig.
// See https://dev.mysql.com/doc/refman/5.7/en/string-functions.html
func (b *builtinLengthSig) evalInt(row types.Row) (int64, bool, error) {
	val, isNull, err := b.args[0].EvalString(b.ctx, row)
	if isNull || err != nil {
		return 0, isNull, errors.Trace(err)
	}
	return int64(len([]byte(val))), false, nil
}

// 我们作为例子添加的
type udfLengthFunctionClass struct{
	baseFunctionClass
}

func (c *udfLengthFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, errors.Trace(err)
	}
	bf := newBaseBuiltinFuncWithTp(ctx, args, types.ETInt, types.ETString)
	bf.tp.Flen = 10
	sig := &builtinUdfLengthSig{bf}
	return sig, nil
}

type builtinUdfLengthSig struct {
	baseBuiltinFunc
}

func (b *builtinUdfLengthSig) Clone() builtinFunc {
	newSig := &builtinUdfLengthSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinUdfLengthSig) evalInt(row types.Row) (int64, bool, error) {
	val, isNull, err := b.args[0].EvalString(b.ctx, row)
	if isNull || err != nil {
		return 0, isNull, errors.Trace(err)
	}
	return int64(len([]byte(val)))+1000, false, nil
}

```

## 在builtin.go中注册

```go
// expression/builtin.go

// funcs holds all registered builtin functions.
var funcs = map[string]functionClass{
	// ...
	ast.Length:          &lengthFunctionClass{baseFunctionClass{ast.Length, 1, 1}},
	ast.UdfLength:      &udfLengthFunctionClass{baseFunctionClass{ast.UdfLength, 1, 1}},
	// ...
	
}
```

## 重新编译运行tidb可以看到udf已经生效

```
MySQL [test]> select runoob_id,length(runoob_author) ,udf_length(runoob_author) from tbl_a where udf_length(runoob_author) >10;
+-----------+-----------------------+----------------------------+
| runoob_id | length(runoob_author) | udf_length(runoob_author) |
+-----------+-----------------------+----------------------------+
|         1 |                    12 |                       1012 |
|         2 |                    12 |                       1012 |
|         3 |                    12 |                       1012 |
+-----------+-----------------------+----------------------------+
3 rows in set (7.16 sec)

MySQL [test]> select runoob_id,length(runoob_author) ,udf_length(runoob_author) from tbl_a where udf_length(runoob_author) <10;
Empty set (3.48 sec)
```
## tidb官方提供的教程

[十分钟成为 TiDB Contributor 系列 | 添加內建函数](https://zhuanlan.zhihu.com/p/25782905)

[十分钟成为 Contributor 系列 | 重构内建函数进度报](https://zhuanlan.zhihu.com/p/27928453)

