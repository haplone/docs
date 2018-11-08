pub trait Messenger{
	fn send(&self,msg: &str);
}

pub struct LimitTracker<'a,T: 'a + Messenger>{
	messenger: &'a T,
	value: usize,
	max: usize,
}

impl<'a,T> LimitTracker<'a,T>	
	where T: Messenger{
	pub fn new(messenger: &T,max: usize ) -> LimitTracker<T>{
		LimitTracker{messenger,value: 0,max,}
	}

	pub fn set_value(&mut self,value: usize){
		self.value = value;
		
		let percentage_of_max = self.value as f64 / self.max as f64;

		if percentage_of_max >= 0.75 && percentage_of_max<0.9{
			self.messenger.send("Warning: You've used up over 75% of your quota!");
		} else if percentage_of_max>=0.9 && percentage_of_max < 1.0 {
			self.messenger.send("Urgent warning:You've used up over 90% of your quota!");
		} else if percentage_of_max <=1.0{
			self.messenger.send("Error: You are over your quota!");
		}
	}
}

#[cfg(test)]
mod tests {
	use super::*;
	use std::cell::RefCell;
	
	struct MockMessenger{
		sent_messengers: RefCell<Vec<String>>,
	}
	
	impl MockMessenger{
		fn new() -> MockMessenger{
			MockMessenger{ sent_messengers: RefCell::new(vec![])}
		}
	}

	impl Messenger for MockMessenger{
		fn send(&self,messenge: &str){
			self.sent_messengers.borrow_mut().push(String::from(messenge));
			//let mut a = self.sent_messengers.borrow_mut();
			//let mut b = self.sent_messengers.borrow_mut();
			//a.push("a".to_string());
			//b.push("b".to_string());
			//let c = self.sent_messengers.borrow();
			//let d = self.sent_messengers.borrow();
		}
	}
	
	#[test]
	fn it_sends_an_over75(){
		let mock_messenger = MockMessenger::new();
		let mut limit_tracker = LimitTracker::new(&mock_messenger,100);

		limit_tracker.set_value(80);

		assert_eq!(mock_messenger.sent_messengers.borrow().len(),1);
	}
}
