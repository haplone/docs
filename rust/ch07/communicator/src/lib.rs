#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        assert_eq!(2 + 2, 4);
    }
}

mod network{
	fn connect(){
		println!("network connect");
	}
	mod server{
		fn connect(){
			println!("network server connect");
		}
	}
}

mod client{
	fn connect(){
		println!("client connect");
	}
}
