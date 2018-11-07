fn main() {
    println!("Hello, world!");

	let v1:Vec<i32> = vec![1,2,3];

	let v2: Vec<i32>= v1.iter().map(|x| x+1).collect();
	
	println!("v1 is {:?}",v1);
	println!("v2 is {:?}",v2);
}
