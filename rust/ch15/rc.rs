enum List{	
	Cons(i32,Rc<List>),
	Nil,
}

use List::{Cons,Nil};
use std::rc::Rc;

fn main(){
	let a = Rc::new(Cons(5,
		Rc::new(Cons(10,
			Rc::new(Nil)))));
	print(&String::from("after a"),&a);
	let b = Cons(3,Rc::clone(&a));
	print(&String::from("after b"),&a);
	{
	let c = Cons(4,Rc::clone(&a));
	print(&String::from("after c"),&a);
	}

	
	print(&String::from("after c gone "),&a);
}

fn print(tag: &str,a: &Rc<List>){
	println!("count {} is {}",tag,Rc::strong_count(&a));
}
