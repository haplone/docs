use std::ops::Deref;

fn main(){
	let x =5;
	let y = &x;

	assert_eq!(5,x);
	assert_eq!(5,*y);

	let yy = Box::new(x);
	assert_eq!(5,*yy);

	let yyy = MyBox::new(x);
	assert_eq!(5,*yyy);

	let name = MyBox::new(String::from("Rust"));
	hello(&name);
}

struct MyBox<T>(T);

impl<T> MyBox<T>{
	fn new(x:T) -> MyBox<T>{
		MyBox(x)
	}
}

impl<T> Deref for MyBox<T>{
	type Target = T;
	fn deref(&self) -> &T{
		&self.0
	}
}

fn hello(name: &str){
	println!("Hello, {}!",name);
}
