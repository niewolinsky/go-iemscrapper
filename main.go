package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bennyscetbun/jsongo"
	"github.com/gocolly/colly/v2"
)

type Iem struct {
	Name                  string `json:"name"`
	Price                 string `json:"price"`
	Price_before_discount string `json:"price_before_discount"`
	Is_unreleased         bool   `json:"is_unreleased"`
}

type Metadata struct {
	Date    string `json:"date"`
	Website string `json:"website"`
	Id      int    `json:"id"`
}

func main() {
	clt := colly.NewCollector()
	URL := "https://hifigo.com/collections/in-ear?page=16&sort_by=price-ascending"

	time_now := time.Now()
	time_formatted := time_now.Format("2006-01-02 15:04:05")

	metadata := Metadata{time_formatted, URL, 1}

	iem_list := []Iem{}

	clt.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	clt.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	clt.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	clt.OnHTML(".productitem--info", func(e *colly.HTMLElement) {
		title := e.ChildText(".productitem--title")
		is_unreleased := false

		price_original_arr := e.ChildTexts(".price--main .money")
		if len(price_original_arr) == 0 {
			price_original_arr = append(price_original_arr, "0.00$")
			is_unreleased = true
		}
		price_original := price_original_arr[0]

		price_discounted_arr := e.ChildTexts(".price--compare-at .money")
		if len(price_discounted_arr) == 0 {
			price_discounted_arr = append(price_discounted_arr, "0.00$")
			is_unreleased = true
		}
		price_discounted := price_discounted_arr[0]

		if price_discounted == "" {
			price_discounted = price_original
		}

		iem_list = append(iem_list, Iem{title, price_original, price_discounted, is_unreleased})
	})

	clt.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		urlSplit := strings.SplitAfter(e.Request.URL.String(), "/")
		baseUrl := urlSplit[0] + urlSplit[1] + urlSplit[2]
		nextUrl := baseUrl + e.ChildAttr("a", "href")
		e.Request.Visit(nextUrl)
	})

	clt.Visit(URL)

	root := jsongo.Node{}
	root.Map("iems").Val(iem_list)
	root.Map("metadata").Val(metadata)
	iemscrapper_data, _ := json.MarshalIndent(&root, "", "  ")

	err := os.WriteFile("iemscrapper_data.json", iemscrapper_data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
