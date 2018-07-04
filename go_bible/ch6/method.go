package main

import (
	"fmt"
)

type Book struct{
	name string
}

func (b Book) printPrice(){
	b.name=b.name+" p"
	fmt.Println(&b)
}

func main(){
	b := Book{"Java"}
	fmt.Println(&b)
	b.printPrice()
	fmt.Println(&b)
}
