package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/haplone/boltdbmvp/html"
	"github.com/haplone/boltdbmvp/model"
	"log"
	"os"
	"time"
	"sync"
)

const ResultFileName = "cars.json"

var targetSize = 2
var currSize = 0
var mu = sync.Mutex{}

/**
遇到的问题：
1. json反序列化时，别名使用没有生效
*/

func main() {
	var carCh = make(chan []model.Car, 10000)
	var urlCh = make(chan string, 10000)

	cities := model.GetCites()
	//cities = cities[:5]
	targetSize = len(cities)

	for _, c := range cities {
		url := fmt.Sprintf(html.StartUrl, c.Pinyin)
		go Fetch(url, carCh, urlCh)
		time.Sleep(time.Millisecond * 50)
	}

	go func() {
		for {
			select {
			case u, ok := <-urlCh:
				log.Printf("consumes url %s", u)
				if ok {
					Fetch(u, carCh, urlCh)
					time.Sleep(50 * time.Millisecond)
				}
			default:
				time.Sleep(time.Microsecond * 10)
			}
		}
	}()

	f, err := os.Create(ResultFileName)
	Check(err)
	w := bufio.NewWriter(f)

	for {
		cs := <-carCh
		if cs != nil && len(cs) > 0 {
			for _, c := range cs {
				Write(w, c)
				//d ,_:= json.Marshal(c)
				//w.WriteString(string(d))
			}
		} else {
			log.Printf("we have write all datas")
			break
		}
	}

	defer w.Flush()
	defer f.Close()
	log.Printf("we have write all datas !!!")
}

func Fetch(url string, carCh chan []model.Car, urlCh chan string) {
	doc := html.FetchHtml(html.Host + url)
	h := html.Html{D: doc}
	currPageNum := h.GetCurrentPage()
	nextPageUrl, _ := h.GetNextPageUrl()

	log.Printf("fetch %s \n", url)
	if currPageNum > 0 && currPageNum <= html.MaxPageSize {
		cars := h.GetCars()

		if cars != nil && len(cars) > 0 {
			carCh <- cars
			log.Printf("we got %d car info\n", len(cars))
		} else {

		}

		if nextPageUrl != "" {
			log.Printf("we got new url %s", nextPageUrl)
			urlCh <- nextPageUrl
		}
	} else {
		mu.Lock()
		currSize += 1
		if currSize >= targetSize {
			carCh <- nil
		}
		mu.Unlock()
	}
}

func Write(w *bufio.Writer, car model.Car) {
	json, err := json.Marshal(car)
	if err != nil {
		log.Printf(" we have problem to convert car (%s) to json", car)
	}
	//log.Println(string(json))
	w.Write(json)
	w.WriteString("\n")
	w.Flush()
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
