use std::thread;
use std::time::Duration;

fn simulated_expensive_calculation(intensity: u32) -> u32{
	println!("Calculating slowly ...");
	thread::sleep(Duration::from_secs(1));
	intensity
}

struct Cacher<T>
	where T: Fn(u32) -> u32
{
	calculation: T,
	value: Option<u32>,
}	

impl<T> Cacher<T>
	where T: Fn(u32) -> u32
{
	fn new(calculation: T) -> Cacher<T>{
		Cacher{
			calculation,
			value: None,
		}
	}
	fn value(&mut self,arg:u32) -> u32{
		match self.value{
			Some(v) => v,
			None => {
				let v = (self.calculation)(arg);
				self.value = Some(v);
				v
			},
		}
	}
}

fn generate_workout(intensity: u32,random_number: u32){
	let mut expensive_closure = Cacher::new(|num| {
		println!("Calculating slowly ...");
		thread::sleep(Duration::from_secs(1));
		intensity
	});

	if intensity < 25 {
		println!("Today , do {} puships!",expensive_closure.value(intensity));
		println!("Next, do {} situps",expensive_closure.value(intensity));
	} else {
		if random_number == 3 {
			println!("Take a break today! Remeber to stay hudrated!");
		} else {
			println!("Today, run for {} minutes!",expensive_closure.value(intensity));
		}
	}
}

fn main() {
	let simulated_user_specified_value = 10;
	let simulated_random_number = 7;

	generate_workout(simulated_user_specified_value,simulated_random_number);
}
