

internal/client/range_lookup.go

RangeLookup is used to look up RangeDescriptors - a RangeDescriptor is a
metadata structure which describes the key range and replica locations of a
distinct range in the cluster. They map the logical keyspace in cockroach to
its physical replicas, allowing a node to send requests for a certain key to
the replicas that contain that key.

RangeLookup用于查找RangeDescriptors  -  RangeDescriptor是一个
 元数据结构，它描述了a的关键字范围和副本位置
 在集群中有明显的范围。他们将蟑螂中的逻辑密钥空间映射到
 其物理副本，允许节点发送某个密钥的请求
 包含该密钥的副本。


RangeDescriptors are stored as values in the cockroach cluster's key-value
store. However, they are always stored using special "Range Metadata keys",
which are "ordinary" keys with a special prefix prepended. The Range Metadata
Key for an ordinary key can be generated with the `keys.RangeMetaKey(key)`
function. The RangeDescriptor for the range which contains a given key can be
retrieved by generating its Range Metadata Key and scanning from the key
forwards until the first RangeDescriptor is found. This is what this function
does with the provided key.

RangeDescriptors作为蟑螂簇的键值中的值存储
 商店。但是，它们始终使用特殊的“范围元数据键”进行存储，
 这是“普通”密钥，前面加了一个特殊的前缀。范围元数据
 普通密钥的密钥可以用`keys.RangeMetaKey（key）`生成
 功能。包含给定键的范围的RangeDescriptor可以是
 通过生成其“范围元数据密钥”并从密钥扫描来进行检索
 转发直到找到第一个RangeDescriptor。这是这个功能
 使用提供的密钥。

Note that the Range Metadata Key sent as the StartKey of the lookup scan is
NOT the key at which the desired RangeDescriptor is stored. Instead, this
method returns the RangeDescriptor stored at the _lowest_ existing key which
is _greater_ than the given key. The returned RangeDescriptor will thus
contain the ordinary key which was provided to this function.

请注意，作为查找扫描的StartKey发送的范围元数据密钥是
 不是所需的RangeDescriptor存储的关键。相反，这
 方法返回存储在 _lowest_ 现有密钥中的RangeDescriptor
 比给定的关键字大于_greater_。返回的RangeDescriptor将如此
 包含提供给该功能的普通密钥。

The "Range Metadata Key" for a range is built by appending the end key of the
range to the respective meta prefix.


一个范围的“范围元数据键”是通过附加的结束键构建的
 范围到相应的元前缀。

It is often useful to think of Cockroach's ranges as existing in a three
level tree:
```
           [/meta1/,/meta1/max)   <-- always one range, gossipped, start here!
                    |
         -----------------------
         |                     |
 [/meta2/,/meta2/m)   [/meta2/m,/meta2/max)
         |                     |
     ---------             ---------
     |       |             |       |
   [a,g)   [g,m)         [m,s)   [s,max)   <- user data
```
In this analogy, each node (range) contains a number of RangeDescriptors, and
these descriptors act as pointers to the location of its children. So given a
key we want to find, we can use the tree structure to find it, starting at
the tree's root (meta1). But starting at the root, how do we know which
pointer to follow? This is where RangeMetaKey comes into play - it turns a
key in one range into a meta key in its parent range. Then, when looking at
its parent range, we know that the descriptor we want is the first descriptor
to the right of this meta key in the parent's ordered set of keys.

在这个类比中，每个节点（范围）都包含一些RangeDescriptor，并且
这些描述符作为指向其子女位置的指针。所以给了一个
我们想找到的关键，我们可以使用树结构来找到它，从开始
树的根（meta1）。但从根开始，我们怎么知道哪一个
要遵循的指针？这就是RangeMetaKey发挥作用的地方 - 它变成了一个
在一个范围内键入其父范围内的一个元键。然后，当看着
它的父范围，我们知道我们想要的描述符是第一个描述符
在父键的有序集合中的这个元键的右侧。

Let's look at a few examples that demonstrate how RangeLookup performs this
task of finding a user RangeDescriptors from cached meta2 descriptors:

我们来看几个演示RangeLookup如何执行此操作的示例
从缓存的meta2描述符中查找用户RangeDescriptors的任务：

Ex. 1:
 Meta2 Ranges: [/meta2/a,  /meta2/z)
 User  Ranges: [a, f) [f, p), [p, z)
 1.a: RangeLookup(key=f)
  In this case, we want to look up the range descriptor for the range [f, p)
  because "f" is in that range. Remember that this descriptor will be stored
  at "/meta2/p". Of course, when we're performing the RangeLookup, we don't
  actually know what the bounds of this range are or where exactly it's
  stored (that's what we're looking up!), so all we have to go off of is the
  lookup key. So, we first determine the meta key for the lookup key using
  RangeMetaKey, which is simply "/meta2/f". We then construct the scan bounds
  for this key using MetaScanBounds. This scan bound will be
  [/meta2/f.Next(),/meta2/max). The reason that this scan doesn't start at
  "/meta2/f" is because if this key is the start key of a range (like it is
  in this example!), the previous range descriptor will be stored at that
  key. We then issue a forward ScanRequest over this range. Since we're
  assuming we already cached the meta2 range that contains this span of keys,
  we send the request directly to that range's replica (if we didn't have
  this cached, the process would recurse to lookup the meta2 range
  descriptor). We then find that the first KV pair we see during the scan is
  at "/meta2/p". This is our desired range descriptor.

  Range -> MetaScanBounds -> ScanRequest

  在这种情况下，我们想查找有“f” 的range[f，p）的range描述符。请记住，此描述符将被存储
  在“/meta2/p”。当然，当我们执行RangeLookup时，实际上我们不
  知道这个range的界限是什么或者它到底存储在哪里
  （这就是我们正在查找的！），所以我们必须离开的是
  查找key。我们首先使用RangeMetaKey查找键来确定meta key
  ，就是“/meta2/f”。为了这个key，我们使用MetaScanBounds构造扫描范围
  。这个扫描界限将会是
  [/meta2/f.Next(),/meta2/max）。
  不能从“/meta2/f”开始扫描，是因为如果这个key是range开始的key（就像这样
  在这个例子中！），前一个range的描述符会存储在那个key上。
  然后，我们在这个range内发布一个前向ScanRequest。因为我们是
  假设我们已经缓存了包含这key范围的meta2 range，
  我们直接将请求发送到该range的副本（如果我们没有缓存，该过程会递归查找meta2 range
  描述符）。然后我们发现我们在扫描过程中看到的第一个KV对
  在“/meta2/p”。这是我们想要的范围描述符。

 1.b: RangeLookup(key=m)
  This case is similar. We construct a scan for this key "m" from
  [/meta2/m.Next(),/meta2/max) and everything works the same as before.
 1.b: RangeLookup(key=p)
  Here, we're looking for the descriptor for the range [p, z), because key "p"
  is included in that range, but not [f, p). We scan with bounds of
  [/meta2/p.Next(),/meta2/max) and everything works as expected.

Ex. 2:
 Meta2 Ranges: [/meta2/a, /meta2/m) [/meta2/m, /meta2/z)
 User  Ranges: [a, f)           [f, p),           [p, z)
 2.a: RangeLookup(key=n)
  In this case, we want to look up the range descriptor for the range [f, p)
  because "n" is in that range. Remember that this descriptor will be stored
  at "/meta2/p", which in this case is on the second meta2 range. So, we
  construct the scan bounds of [/meta2/n.Next(),/meta2/max), send this scan
  to the second meta2 range, and find that the first descriptor found is the
  desired descriptor.
 2.b: RangeLookup(key=g)
  This is where things get a little tricky. As usual, we construct scan
  bounds of [/meta2/g.Next(),/meta2/max). However, this scan will be routed
  to the first meta2 range. It will scan forward and notice that no
  descriptors are stored between [/meta2/g.Next(),/meta2/m). We then rely on
  DistSender to continue this scan onto the next meta2 range since the result
  from the first meta2 range will be empty. Once on the next meta2 range,
  we'll find the desired descriptor at "/meta2/p".

Ex. 3:
 Meta2 Ranges: [/meta2/a, /meta2/m)  [/meta2/m, /meta2/z)
 User  Ranges: [a, f)        [f, m), [m,s)         [p, z)
 3.a: RangeLookup(key=g)
  This is a little confusing, but actually behaves the exact same way at 2.b.
  Notice that the descriptor for [f, m) is actually stored on the second
  meta2 range! So the lookup scan will start on the first meta2 range and
  continue onto the second before finding the desired descriptor at /meta2/m.
  This is an unfortunate result of us storing RangeDescriptors at
  RangeMetaKey(desc.EndKey) instead of RangeMetaKey(desc.StartKey) even
  though our ranges are [inclusive,exclusive). Still everything works if we
  let DistSender do its job when scanning over the meta2 range.

  See #16266 and #17565 for further discussion. Notably, it is not possible
  to pick meta2 boundaries such that we will never run into this issue. The
  only way to avoid this completely would be to store RangeDescriptors at
  RangeMetaKey(desc.StartKey) and only allow meta2 split boundaries at
  RangeMetaKey(existingSplitBoundary)


Lookups for range metadata keys usually want to read inconsistently, but some
callers need a consistent result; both are supported be specifying the
ReadConsistencyType. If the lookup is consistent, the Sender provided should
be a TxnCoordSender.

This method has an important optimization if the prefetchNum arg is larger
than 0: instead of just returning the request RangeDescriptor, it also
returns a slice of additional range descriptors immediately consecutive to
the desired RangeDescriptor. This is intended to serve as a sort of caching
pre-fetch, so that nodes can aggressively cache RangeDescriptors which are
likely to be desired by their current workload. The prefetchReverse flag
specifies whether descriptors are prefetched in descending or ascending
order.
