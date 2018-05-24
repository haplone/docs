package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	d1 := []byte("hello\ngo\n")
	err := ioutil.WriteFile("/tmp/data1", d1, 0644)
	check(err)

	f, err := os.Create("/tmp/data2")
	check(err)

	defer f.Close()

	d2 := []byte{115, 111, 109, 101, 10}
	n2, err := f.Write(d2)
	check(err)
	fmt.Printf("write data %d \n", n2)

	n3, err := f.WriteString("writes\n")
	check(err)
	fmt.Printf("write data %d \n", n3)
	f.Sync()

	w := bufio.NewWriter(f)
	n4, err := w.WriteString("buffered \n")
	check(err)
	fmt.Printf("wrote %d bytes \n", n4)

	w.Flush()

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
