package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	ch := make(chan int, 10)
	var idx int = 0

	fmt.Println("start")
	go func() {
		//var size  = 100
		for {
			//i := idx
			fmt.Println("sd")
			log.Println(idx)
			//func(i int){
			//	fmt.Println(i)
			ch <- idx
			//}(i)
			idx += 1

		}
		//for{
		//	if idx <= size {
		//		fmt.Printf("write %d",idx)
		//		ch <- idx
		//		idx = idx+1
		//	//} else {
		//	//	close(ch)
		//	//	return
		//	}
		//}
	}()

	go func() {
		for {
			n := <-ch
			log.Printf("got %d", n)
		}
	}()

	time.Sleep(1 * time.Second)
	log.Println(idx)
}
