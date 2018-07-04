package main

import "fmt"

func main() {
	s := "left"
	t := s
	fmt.Println(&s)
	s += ",right"

	fmt.Println(s)
	fmt.Println(*&s)
	fmt.Println(&s)
	fmt.Println(t)
	fmt.Println(&t)

	const Annonce = `here
	we
	are
`
	fmt.Println(Annonce)

	for i, r := range "Hello,世界" {
		fmt.Printf("%d \t %q \t %d \n", i, r, r)
	}
}
