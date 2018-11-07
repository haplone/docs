use std::fs::File;
use std::io::prelude::*;
use std::error::Error;


pub fn run(config: Config) -> Result<(),Box<Error>>{
	let mut f = File::open(config.filename)?;
	let mut contents = String::new();
	f.read_to_string(&mut contents)?;

	println!("With text:\n{}",contents);
	let lines = search(&config.query,&contents);
	for line in lines{
		println!("-{}",line);
	}
	Ok(())
}

pub fn search<'a>(query: &str,contents: &'a str) -> Vec<&'a str>{
	let mut results = Vec::new();

	for line in contents.lines(){
		if line.contains(query){
			results.push(line);
		}
	}
	results
}

pub struct Config{
	query: String,
	filename: String,
}

impl Config{
	pub fn new(args: &[String]) -> Result<Config,&'static str>{
		if args.len() < 3{
			return Err("not enough arguments");
		}

		let query = args[1].clone();
		let filename = args[2].clone();
		println!("Searching for {} in file {}",query,filename);
		Ok(Config{query,filename})
	}
}

#[cfg(test)]
mod test {
	use super::*;

	#[test]
	fn one_result(){
		let query = "duct";
		let contents= "\
Rust:
safe, fast, productive.
Pick three.";
		assert_eq!(
			vec!["safe, fast, productive."],
			search(query,contents)
		);
	}
}
