package main

import (
	"github.com/haplone/boltdbmvp/html"
	"github.com/haplone/boltdbmvp/model"
	//"sync"
	"time"
	"log"
	"bufio"
	"encoding/json"
	"os"
)

func main() {
	var seriesCh = make(chan []model.Series, 100)
	var specCh = make(chan []model.CarSpec, 100)

	spf, _ := os.Create(SpecFile)
	//Check(err)
	spw := bufio.NewWriter(spf)
	defer spw.Flush()
	defer spf.Close()

	sef, _ := os.Create(SeriesFile)
	//Check(err)
	sew := bufio.NewWriter(sef)
	defer sew.Flush()
	defer sef.Close()

	//wg := sync.WaitGroup{}
	doc := html.FetchHtml(html.Host + InitUrl)
	h := html.Html{D: doc}
	brands := h.GetBrands()

	for i := 0; i < 10; i++ {
		go GetSpec(seriesCh, specCh, sew)
	}
	go writeSpec(spw,specCh)

	//wg.Add(len(brands))
	idx := 0
	for _, brand := range brands {
		go GetSeries(brand.Url, seriesCh)
		idx ++
		if idx%20 == 0 {
			time.Sleep(1000 * time.Millisecond)
		}
	}

	//wg.Wait()
	time.Sleep(1000*time.Second)
}

func GetSeries(url string, seriesCh chan [] model.Series) {
	log.Printf("we are fetching series %s \n", url)
	seriesDoc := html.FetchHtml(html.Host + url)
	seriesH := html.Html{D: seriesDoc}
	series := seriesH.GetSeries()
	seriesCh <- series
}

func GetSpec(seriesCh chan [] model.Series, specCh chan []model.CarSpec, w *bufio.Writer) {
	for {
		select {
		case series, ok := <-seriesCh:
			if ok {
				for _, s := range series {
					writeSerie(w, s)
					log.Printf("we are fetching specs %s \n", s.Url)
					doc := html.FetchHtml(html.Host + s.Url)
					h := html.Html{D: doc}
					specs := h.GetCarSpecs()
					for _, car := range specs {
						car.SerieId = s.Id
					}
					specCh <- specs
				}
			}
		}
	}

}

func writeSpec(w *bufio.Writer, specsCh chan []model.CarSpec) {
	for {
		select {
		case specs, ok := <-specsCh:
			if ok {
				for _, car := range specs {
					json, err := json.Marshal(car)
					if err != nil {
						log.Printf(" we have problem to convert car (%s) to json", car)
					}
					//log.Println(string(json))
					w.Write(json)
					w.WriteString("\n")
					w.Flush()
				}
			}
		}
	}

}

func writeSerie(w *bufio.Writer, series model.Series) {
	json, err := json.Marshal(series)
	if err != nil {
		log.Printf(" we have problem to convert series (%s) to json", series)
	}
	//log.Println(string(json))
	w.Write(json)
	w.WriteString("\n")
	w.Flush()
}

const (
	InitUrl       = "/AsLeftMenu/As_LeftListNew.ashx?typeId=1%20&brandId=276%20&fctId=0%20&seriesId=0"
	SeriesInitUrl = "/price/brand-40.html"
	SpecFile      = "spec.json"
	SeriesFile    = "series.json"
)
