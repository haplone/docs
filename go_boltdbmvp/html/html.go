package html

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"log"
	"strings"
	"github.com/haplone/boltdbmvp/model"
	"regexp"
	"net/http"
	"github.com/go-crawler/car-prices/fake"
	"github.com/axgle/mahonia"
)

const (
	Host        = "https://car.autohome.com.cn"
	StartUrl    = "/2sc/%s/a0_0msdgscncgpi1ltocsp1exb4/"
	MaxPageSize = 99
)

var (
	compileNumber = regexp.MustCompile("[0-9]")
)

/**
	debug 这边如何改?
 */
type Html struct {
	D *goquery.Document
}

func (h *Html) GetCityName() string {
	return h.D.Find(".citycont .fn-left").Text()
}

func (h *Html) GetNextPageUrl() (v string, exists bool) {
	return h.D.Find(".page .page-item-next").Attr("href")
}

func (h *Html) GetCurrentPage() (page int) {
	pageS := h.D.Find(".page .current").Text()

	if pageS != "" {
		var err error
		page, err = strconv.Atoi(pageS)
		if err != nil {
			log.Printf("spiders.GetCurrentPage err: %v", err)
		}
	}
	return page
}

func (h *Html) GetBrands() []model.Brand {
	var brands []model.Brand
	h.D.Find(" ul li").Each(func(i int, selection *goquery.Selection) {
		id, _ := selection.Attr("id")
		url, _ := selection.Find("h3 a").Attr("href")
		name := selection.Find("h3 a").Text()

		var brand = model.Brand{
			Id:   id,
			Url:  url,
			Name: name,
		}
		log.Println("brand : ", brand)
		brands = append(brands, brand)
	})
	return brands
}

func (h *Html) GetSeries() (s []model.Series) {
	h.D.Find("#brandtab-1 div.list-cont-bg").Each(func(i int, selection *goquery.Selection) {
		id, _ := selection.Find(".main-title").Attr("id")
		imgUrl, _ := selection.Find(".list-cont-img img").Attr("src")
		name := selection.Find(".main-title").Text()
		url, _ := selection.Find(".main-title a").Attr("href")
		score := selection.Find(".score-number").Text()
		price := selection.Find(".lever-price").Text()
		var details []string
		selection.Find(".lever-ul li").Each(func(i int, qs *goquery.Selection) {
			details = append(details, qs.Text())
		})

		series := model.Series{
			Id:       id,
			ImageUrl: imgUrl,
			Name:     name,
			Url:      url,
			Score:    score,
			Price:    price,
			Details:  details,
		}
		log.Println(series)
		s = append(s, series)
	})
	log.Printf("we got %d series \n", len(s))
	return
}

func (h *Html) GetCarSpecs() (cs []model.CarSpec) {
	h.D.Find(".interval01-list li").Each(func(i int, qs *goquery.Selection) {
		id, _ := qs.Attr("data-value")
		name := qs.Find(".interval01-list-cars-infor p a").Text()
		var details []string

		qs.Find(".interval01-list-cars-infor p span").Each(func(i int, s *goquery.Selection) {
			details = append(details, s.Text())
		})
		price := qs.Find(".interval01-list-guidance").Text()

		car := model.CarSpec{
			Id:      id,
			Name:    name,
			Price:   price,
			Details: details,
		}
		cs = append(cs, car)
	})
	return
}

func (h *Html) GetCars() (cars []model.Car) {
	cityName := h.GetCityName()
	h.D.Find(".piclist ul li:not(.line)").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find(".title a").Text()
		price := selection.Find(".detail .detail-r").Find(".colf8").Text()
		kilometer := selection.Find(".detail .detail-l").Find("p").Eq(0).Text()
		year := selection.Find(".detail .detail-l").Find("p").Eq(1).Text()

		dealerId, _ := selection.Attr("dealerid")
		infoId, _ := selection.Attr("infoid")
		// 数据处理
		kilometer = strings.Join(compileNumber.FindAllString(kilometer, -1), "")
		year = strings.Join(compileNumber.FindAllString(strings.TrimSpace(year), -1), "")
		priceS, _ := strconv.ParseFloat(price, 64)
		kilometerS, _ := strconv.ParseFloat(kilometer, 64)
		yearS, _ := strconv.Atoi(year)

		car := model.Car{
			CityName:  cityName,
			Title:     title,
			Price:     priceS,
			Kilometer: kilometerS,
			Year:      yearS,
			DealerId:  dealerId,
			InfoId:    infoId,
		}
		cars = append(cars, car)
	})

	return cars
}

func FetchHtml(url string) *goquery.Document {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("http.NewRequest err: %v", err)
	}

	req.Header.Add("User-Agent", fake.GetUserAgent())
	req.Header.Add("Referer", Host)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Printf("client.Do err: %v", err)
	}

	mah := mahonia.NewDecoder("gbk")
	body := mah.NewReader(resp.Body)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Printf("Downloader.Get err: %v", err)
	}

	return doc
}
