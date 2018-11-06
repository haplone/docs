extern crate rand;

use std::io;
use std::cmp::Ordering;
use rand::Rng;

pub struct Game{
	secret_number: u32,
}

impl Game{
	pub fn new() -> Game {
		let num = rand::thread_rng().gen_range(1,101);
		println!("secret number is {}",num);
		Game{ secret_number: num }
	}
	pub fn cmp(&self,n: &Guess) -> bool {
		match self.secret_number.cmp(&n.value) {
			Ordering::Less => {
				println!("Too big!");
				false
			},
			Ordering::Greater => {
				println!("Too small!");
				false
			},
			Ordering::Equal => {
				println!("You win!");
				true
			},
		}
	}
	pub fn start(&self) {
		println!(" Guess the number!");
	}	
}

pub struct Guess {
	value: u32
}

impl Guess {
	pub fn new(guess: &String) -> Option<Guess>{
		let guess: u32 = match guess.trim().parse(){
			Ok(num) => num,
			Err(_) => return None
		};
		if guess <1 || guess > 100{
			return None	
		}

		Some(Guess{value: guess})
	}
}

fn main(){
	let game = Game::new();

	game.start();

	loop{
		println!("Please input your guess.");
		let mut guess = String::new();
		io::stdin().read_line(&mut guess).expect("Failed to read line");

		let guess = Guess::new(&guess);

		match guess {
			Some(g) => {
				println!("You guessed: {}",g.value);
				
				if game.cmp(&g){
					break;
				}
			},
			None => {
				println!("Please input the number between 1 and 100");
				continue
			},
		}
		
	}
}	
