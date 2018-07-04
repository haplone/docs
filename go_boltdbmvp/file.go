package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"log"
	"io"
	"encoding/json"
	"github.com/haplone/boltdbmvp/model"
)

func main() {
	var carCh = make(chan model.Car, 100)

	go read(carCh)

	getCity(carCh)
}

func getCity(carCh chan model.Car) {
	var cities = make(map[string]int)
	defer func() {
		log.Println(cities)
		var idx = 0
		for _, c := range cities {
			idx += c
		}
		log.Printf("we got %d car infos \n",idx)
	}()

	for {
		select {
		case c, ok := <-carCh:
			//log.Println("city name", c.CityName != "")
			if ok && c.CityName != "" {
				count := cities[c.CityName]
				//log.Printf(" %s has %d [ %s]", c.CityName, count, c.Title)
				cities[c.CityName] = count + 1
			} else {
				return
			}
		}
	}
}

func read(carCh chan model.Car) {
	f, err := os.Open("used_cars.json")
	if err != nil {
		log.Fatal(err)
	}

	b := bufio.NewReader(f)

	var idx = 0
	defer f.Close()
	for {
		l, _, err := b.ReadLine()
		if err != io.EOF {
			idx += 1
			var c model.Car
			json.Unmarshal(l, &c)
			//log.Println(c)
			carCh <- c
		} else {
			carCh <- model.Car{CityName: ""}
			log.Println("we've read all file ")
			break
		}
	}
	//defer func() {
	//	close(carCh)
	//}()

	defer log.Printf("we got %d line \n", idx)
}

func write() {
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
