package main

import (
	"time"
	"log"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	wg.Add(10)

	go DoBiz(&wg)
	wg.Wait()
	Wait()
}

func Wait() {
	log.Println("============= done")
}
func DoBiz(wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()
	count := 0
	for {
		c := <-ticker.C
		log.Printf(" ticker %d %v \r\n", count, c)
		count += 1
		wg.Done()
	}
}
