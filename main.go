package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Iem struct {
	name             string
	price_original   string
	price_discounted string
	is_unreleased    bool
}

func main() {
	clt := colly.NewCollector()
	iem_ls := []Iem{}

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

		iem_ls = append(iem_ls, Iem{title, price_original, price_discounted, is_unreleased})
	})

	clt.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		urlSplit := strings.SplitAfter(e.Request.URL.String(), "/")
		baseUrl := urlSplit[0] + urlSplit[1] + urlSplit[2]
		nextUrl := baseUrl + e.ChildAttr("a", "href")
		e.Request.Visit(nextUrl)
	})

	clt.Visit("https://hifigo.com/collections/in-ear?page=16&sort_by=price-ascending")
	for _, el := range iem_ls {
		fmt.Println(el)
	}
}
