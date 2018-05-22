package main

import "fmt"

func main() {
	var a [3]int
	fmt.Println(a[0])
	fmt.Println(a[len(a)-1])

	for i, v := range a {
		fmt.Println("%d %d\n", i, v)
	}

	var q [3]int = [3]int{1, 2, 3}
	var r [3]int = [3]int{1, 2}
	print(q)
	print(r)
}

func print(a [3]int) {
	for i, v := range a {
		fmt.Println("%d %d\n", i, v)
	}
}
