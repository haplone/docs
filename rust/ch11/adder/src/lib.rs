#[derive(Debug)]
pub struct Rectangle{
	length: u32,
	width: u32,
}

impl Rectangle{
	pub fn can_hold(&self,other: &Rectangle) -> bool {
		self.length>other.length && self.width>other.width
	}
}

#[cfg(test)]
mod tests {
	#[test]
	fn it_works(){
		assert_eq!(2+2,4);
	}
	use super::*;
	
	#[test]
	fn larger_can_hold_smaller(){
		let larger = Rectangle{length: 8,width: 7};
		let smaller = Rectangle{ length: 5, width: 1};
		assert!(larger.can_hold(&smaller));
	}
}
