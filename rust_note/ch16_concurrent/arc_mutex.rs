use std::sync::{Mutex,Arc};
use std::thread;

fn main(){
    let counter = Arc::new(Mutex::new(0));
    let mut handles = vec![];

    for _ in 0..10{
        let c = Arc::clone(&counter);
        let handle = thread::spawn(move||{
            let mut num = c.lock().unwrap();
            *num += 1;
        });

        handles.push(handle);
    }

    for handle in handles {
        handle.join().unwrap();
    }

    println!("Result: {:?}",counter);
    println!("Result: {}",*counter.lock().unwrap());
}
        
