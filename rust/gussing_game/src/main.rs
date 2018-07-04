use std::io;


fn main() {
    println!("Guess the number!");
    println!("Please input your gues.");

    let mut guess = String::new();

    io::stdin().read_line(&mut guess)
      .expect("Fialed to read line");

    println!("Your guess : {}",guess);
}
