package main

import "fmt"
import "time"

const noDelay time.Duration = 2
const timeout time.Duration = 0

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

func main() {
	fmt.Printf("%T %[1]v \n", timeout)

	fmt.Println(Monday)
}
