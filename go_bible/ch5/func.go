package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	defer trace("func test")()
	fmt.Println(strings.Map(add1, "HAL-9000"))
	fmt.Println(strings.Map(add1, "VMS"))
	fmt.Println(strings.Map(add1, "Admin"))

	var f = squreas()
	fmt.Println(f())
	fmt.Println(f())
	fmt.Println(f())

	defer fmt.Println("bye bye")
	fmt.Println(sum(1, 2, 3, 4, 5))
	nums := []int{1, 2, 3, 4, 5}

	fmt.Println(sum(nums...))
}

func trace(msg string) func() {
	start := time.Now()
	log.Printf("enter %s", msg)
	return func() {
		log.Printf("exit %s (%s)", msg, time.Since(start))
	}
}

func add1(r rune) rune { return r + 1 }

func squreas() func() int {
	var x int
	return func() int {
		x++
		return x * x
	}
}

func sum(vals ...int) int {
	total := 0
	for _, val := range vals {
		total += val
	}
	return total
}

/*
func g(x int) func(y int) func(z int) a{
	return func(y int) func(z int) a{
		return func(z int){
			return x*y*z*10
		}
	}
}
*/
