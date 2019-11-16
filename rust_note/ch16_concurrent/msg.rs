use std::thread;
use std::sync::mpsc;

fn main(){
    let (tx,rx) = mpsc::channel();

    let _handle = thread::spawn(move||{
        for i in 1..10{
            let val = format!("hi {}",i);
            tx.send(val).unwrap();
        }
    });

    //handle.join().unwrap();
    
    for _i in 1..10{
        let received = rx.recv().unwrap();
        println!("Got {}",received);
    }
}
