fn main() {
	let mut x = 5;
	println!("The value of x is {}",x);
	x= 6;
	println!("The value of x is {}", x);

	let y =5;
	let y= y + 1;
	let y = y*2;
	println!("The value of y is {}", y);

	let spaces = "         ";
	let spaces = spaces.len();
	println!("The values of spaces is {}", spaces);

//	let mut spaces2 = "        ";
//	spaces2 = spaces2.len();

	let guess : u32 = "42".parse().expect("Not a number!");	

	println!("The value of guess is : {}", guess);

	let tup :(i32,f64,u8) = (500,6.4,1);

	println!("The value of tup is {:?}", tup);

	let (a,b,c) = tup;
	println!("The value of a is {}, b is {}, c is {}",a,tup.1,tup.2);

	let array = [1,2,3,4,5];
//	let index = 10;
	let index =2 ;
	let element = array[index];
	println!("The value of element is {}", element);


	haha();
}

fn haha(){
	let x=5;
	
	let y = {
		let x =3;
		x+1
	};
	println!("The value of x is {}, y is {}", x,y);
}
