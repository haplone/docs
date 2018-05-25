package model

import (
	"os"
	"io/ioutil"
	"log"
	"encoding/json"
	//"fmt"
)

const (
	Province_city_json_file = "province_city.json"
)

type Province struct {
	Id             int32  `json: Id`
	Name           string
	FirstCharacter string
	Pinyin         string
	City           []City `json: "City"`
}

type City struct {
	Id             int32
	Name           string
	FirstCharacter string
	Pinyin         string
}

type Car struct {
	CityName  string  `json: "city_name"`
	Title     string  `json: "title"`
	Price     float64 `json: "price"`
	Kilometer float64 `json: "kilometer"`
	Year      int     `json: "year"`
	InfoId    string  `json: "info_id"`
	DealerId  string  `json: "dealer_id"`
}

type Brand struct{
	Id string
	Name string
	Url string
}

type Series struct{
	Id string
	ImageUrl string
	Name string
	Url string

}
func GetCites() []City {
	if checkFileIsExists(Province_city_json_file) {
		data, err := ioutil.ReadFile(Province_city_json_file)
		if err != nil {
			log.Fatalf("read province city json file error %s", err)
		}
		var ps []Province
		var cs []City
		json.Unmarshal(data, &ps)

		for _, p := range ps {
			for _, c := range p.City {
				cs = append(cs, c)
			}
		}

		log.Printf(" we load %d provice %d city \n", len(ps), len(cs))
		return cs
	}
	return nil
}

func checkFileIsExists(fn string) (exist bool) {
	exist = true

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		exist = false
	}
	return
}
