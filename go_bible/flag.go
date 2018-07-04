package main

import(
	"fmt"
	"flag"
)

func main(){
	fmt.Println("hello")

	var ip=flag.Int("flagname",1234,"help msg for flagname")
	flag.Parse()

	fmt.Printf("ip: %d",*ip)

}
