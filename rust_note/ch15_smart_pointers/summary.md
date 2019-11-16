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

`Drop` trait 允许我们在值离开作用域时执行一些代码。

```rust
struct CustomSmartPointer {
    data: String,
}

impl Drop for CustomSmartPointer {
    fn drop(&mut self) {
        println!("Dropping CustomSmartPointer with data `{}`!", self.data);
    }
}

fn main() {
    let c = CustomSmartPointer { data: String::from("my stuff") };
    let d = CustomSmartPointer { data: String::from("other stuff") };
    println!("CustomSmartPointers created.");
}
```

使用std::mem::drop提早丢弃值

```rust
fn main() {
    let c = CustomSmartPointer { data: String::from("some data") };
    println!("CustomSmartPointer created.");
    c.drop();
    println!("CustomSmartPointer dropped before the end of main.");
}
```

## `Rc<T>`引用计数智能指针

`Rc<T>` 用于当我们希望在堆上分配一些内存供程序的多个部分读取，且无法在编译时确定程序哪一部分最后结束使用。

```rust
enum List {
    Cons(i32, Rc<List>),
    Nil,
}

use crate::List::{Cons, Nil};
use std::rc::Rc;

fn main() {
    let a = Rc::new(Cons(5, Rc::new(Cons(10, Rc::new(Nil)))));
    let b = Cons(3, Rc::clone(&a));
    let c = Cons(4, Rc::clone(&a));
}
```

克隆Rc<T>会增加引用计数： `Rc::strong_count(&a)`

## `RefCell<T>`和内部可变性模式

内部可变性interior mutability 是rust中的一个设计模式，它允许你即使在有不可变引用时改变数据。内部实现使用`unsafe`。

借用规则适用于运行时。

* `Rc<T>` 允许相同数据有多个所有者`Box<T>` 和 `RefCell<T>` 有单一所有者。
* `Box<T>` 运行在编译时不可变或可变借用检查`Rc<T>` 仅允许在编译时执行不可变借用检查。
* 因`RefCell<T>` 允许在运行时执行可变借用检查，所以我们可以在即便`RefCell<T>`自身是不可变的情况下修改其内部值。

## 引用循环与内存泄漏

Rust的内存安全保证使其难以意外制造永远也不会被清理的内存（内存泄漏 memory leak），但并不是不可能。与在编译时，拒绝数据竞争不同，Rust并不保证完全避免内存泄漏，这意味着内存泄漏在Rust被认为是内存安全的。


```rust
fn main() {
    let a = Rc::new(Cons(5, RefCell::new(Rc::new(Nil))));

    println!("a initial rc count = {}", Rc::strong_count(&a));
    println!("a next item = {:?}", a.tail());

    let b = Rc::new(Cons(10, RefCell::new(Rc::clone(&a))));

    println!("a rc count after b creation = {}", Rc::strong_count(&a));
    println!("b initial rc count = {}", Rc::strong_count(&b));
    println!("b next item = {:?}", b.tail());

    if let Some(link) = a.tail() {
        *link.borrow_mut() = Rc::clone(&b);
    }

    println!("b rc count after changing a = {}", Rc::strong_count(&b));
    println!("a rc count after changing a = {}", Rc::strong_count(&a));

    // 取消如下行的注释来观察引用循环；
    // 这会导致栈溢出
    // println!("a next item = {:?}", a.tail());
}
```

* 将`Rc<T>` 替换为 `Weak<T>`
* 使用`RC::downgrade` 传递`Rc` 实例的引用来创建其值的弱引用weak reference。
* 使用`upgrade` 获取引用`Option<Rc<T>>`

```rust
fn main() {
    let leaf = Rc::new(Node {
        value: 3,
        parent: RefCell::new(Weak::new()),
        children: RefCell::new(vec![]),
    });

    println!(
        "leaf strong = {}, weak = {}",
        Rc::strong_count(&leaf),
        Rc::weak_count(&leaf),
    );

    {
        let branch = Rc::new(Node {
            value: 5,
            parent: RefCell::new(Weak::new()),
            children: RefCell::new(vec![Rc::clone(&leaf)]),
        });

        *leaf.parent.borrow_mut() = Rc::downgrade(&branch);

        println!(
            "branch strong = {}, weak = {}",
            Rc::strong_count(&branch),
            Rc::weak_count(&branch),
        );

        println!(
            "leaf strong = {}, weak = {}",
            Rc::strong_count(&leaf),
            Rc::weak_count(&leaf),
        );
    }

    println!("leaf parent = {:?}", leaf.parent.borrow().upgrade());
    println!(
        "leaf strong = {}, weak = {}",
        Rc::strong_count(&leaf),
        Rc::weak_count(&leaf),
    );
}
```
