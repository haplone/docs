# 智能指针

## 概念

* 指针： 包含内存地址的变量的通用概念。Rust最常见的指针是引用reference。引用以`&`为标志并借用他们所指向的值。
* 智能指针smart pointers： 一类数据结构，表现类似指针，拥有额外的元数据和功能。

在rust中，引用是一类`借用`数据的指针;智能指针`拥有`他们指向的数据

之前的`String`和`Vec<T>`就是智能指针。

智能指针跟常规结构体的显著区别是`Deref` 和 `Drop` trait。

本章主要讲述：

* `Box<T>` 用于堆上分配值
* `Rc<T>` 引用计数类型，其数据可以有多个所有者
* `Ref<T>` 和 `RefMut<T>`，通过`RefCell<T>` 访问，运行时（而非编译时）执行借用规则.
* 内部可变性interior mutability
* 引用循环reference cycles

## `Box<T>` 堆上存储数据，且可确定大小

box允许你将一个值放在堆上而不是栈上留在栈上的则是指向堆数据的指针。box没有性能损失。

box多用于如下场景：
* 一个类型编译时大小未知，又想在确切大小的上下文中使用： 如递归
* 确保大量数据不被拷贝的情况下转移所有权
* 希望拥有一个值且只关心是否实现特定trait，而不是具体类型


### 递归场景：

```rust
enum List {
    Cons(i32, Box<List>),
    Nil,
}

use crate::List::{Cons, Nil};

fn main() {
    let list = Cons(1,
        Box::new(Cons(2,
            Box::new(Cons(3,
                Box::new(Nil))))));
}
```

## 通过`Deref` trait将智能指针当作常规引用处理

* 实现`Deref` trait允许我们重载解引用运算符dereference operator `*`。
* 解引用强制多态deref coercions

```rust

struct MyBox<T>(T);

impl<T> MyBox<T> {
    fn new(x: T) -> MyBox<T> {
        MyBox(x)
    }
}

use std::ops::Deref;

impl<T> Deref for MyBox<T> {
    type Target = T;

    fn deref(&self) -> &T {
        &self.0
    }
}

fn main() {
    let x = 5;
    let y = MyBox::new(x);

    assert_eq!(5, x);
	// *(y.deref())
    assert_eq!(5, *y);
}


```

### 解引用强制多态
```rust

fn hello(name: &str) {
    println!("Hello, {}!", name);
}

fn main() {
    let m = MyBox::new(String::from("Rust"));
    hello(&m);
}
```

如果没有解引用强制多态：
```rust
fn main() {
    let m = MyBox::new(String::from("Rust"));
    hello(&(*m)[..]);
}
```

### 解引用强制多态与可变性交互

* 当`T:Deref<Target=U>` 时 `&T` 到 `&U`
* 当`T:DerefMut<Target=U>'时 `&mut T` 到 `&mut U`
* 当`T:Deref<Target=U>` 时 `&mut T` 到 `&U`


## 使用`Drop` trait运行清理代码


