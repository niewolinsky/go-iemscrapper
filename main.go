package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	clt := colly.NewCollector()
	ls := []string{}

	clt.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	clt.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	clt.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	clt.OnHTML(".productgrid--items", func(e *colly.HTMLElement) {
		ls = append(ls, e.ChildTexts(".productitem--title")...)
	})

	clt.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		urlSplit := strings.SplitAfter(e.Request.URL.String(), "/")
		baseUrl := urlSplit[0] + urlSplit[1] + urlSplit[2]
		nextUrl := baseUrl + e.ChildAttr("a", "href")
		e.Request.Visit(nextUrl)
	})

	clt.Visit("https://hifigo.com/collections/in-ear?sort_by=price-ascending")
	for _, el := range ls {
		x := strings.TrimSpace(el)
		fmt.Println(x)
	}
}
