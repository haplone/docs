struct CustomSmartPointer{
	data: String,
}

impl Drop for CustomSmartPointer{
	fn drop(&mut self){
		println!("Dropping CustomSmartPointer with data :{}",self.data);
	}
}

fn main(){
	let c = CustomSmartPointer{ data: String::from("rust")};
	let d = CustomSmartPointer{ data: String::from("go")};


	let e = CustomSmartPointer{ data: String::from("cargo")};
	println!("created CustomSmartPointers ");
	drop(e);
	println!("leaving");
}
