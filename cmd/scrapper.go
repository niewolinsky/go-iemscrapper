package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/niewolinski/go-iemscrapper/helpers"

	"github.com/gocolly/colly/v2"
)

type Iem struct {
	Name                  string `json:"name" bson:"name"`
	Price                 string `json:"price" bson:"price"`
	Price_before_discount string `json:"price_before_discount" bson:"price_before_discount"`
	Is_unreleased         bool   `json:"is_unreleased" bson:"is_unreleased"`
	Rank                  string `json:"ranking" bson:"ranking"`
}

type Ranking struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
	Rank  string `json:"rank" bson:"rank"`
}

func (app *application) initScrapper(url string) *colly.Collector {
	base_url_trim := helpers.GetBaseUrl(url)

	clt := colly.NewCollector(
		colly.AllowedDomains(base_url_trim),
	)

	clt.OnError(func(_ *colly.Response, err error) {
		fmt.Fprintf(os.Stderr, "error while scraping the website: %v \n", err)
		os.Exit(1)
	})

	clt.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	clt.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	return clt
}

func appendRankings(ranking_list []Ranking, iem_list []Iem) {
	//use goroutine??
	for _, ranking := range ranking_list {
		rank_name_split := strings.Split(ranking.Name, " ")

		for j, iem := range iem_list {
			if iem_list[j].Rank != "" {
				continue
			}

			iem_name_split := strings.Split(iem.Name, " ")
			is_equal := true

			if len(rank_name_split) <= len(iem_name_split) {
				num := len(rank_name_split) - 1
				for i := num; i >= 0; i = i - 1 {
					if !strings.EqualFold(rank_name_split[i], iem_name_split[i]) {
						is_equal = false
						break
					}
				}

				if is_equal {
					iem_list[j].Rank = ranking.Rank
					fmt.Printf("IEM: %v, RANK: %v", iem, ranking)
				}
			}
		}
	}
}

const (
	url_shop    = "https://hifigo.com/collections/in-ear?sort_by=price-ascending"
	url_ranking = "https://crinacle.com/rankings/iems/"
)

func (app *application) scrapData() {
	iem_list := []Iem{}
	ranking_list := []Ranking{}

	clt_shop := app.initScrapper(url_shop)
	clt_ranking := app.initScrapper(url_ranking)

	clt_ranking.OnHTML("#tablepress-4", func(e *colly.HTMLElement) {
		name := e.ChildTexts(".column-3")
		value := e.ChildTexts(".column-2")
		rank := e.ChildTexts(".column-1")

		for i := range name {
			ranking_list = append(ranking_list, Ranking{name[i], value[i], rank[i]})
		}
	})

	clt_shop.OnHTML(".productitem--info", func(e *colly.HTMLElement) {
		title := e.ChildText(".productitem--title")
		is_unreleased := false

		price_original_arr := e.ChildTexts(".price--main .money")
		if len(price_original_arr) == 0 {
			price_original_arr = append(price_original_arr, "$0.00")
			is_unreleased = true
		}
		price_original := price_original_arr[0]

		price_discounted_arr := e.ChildTexts(".price--compare-at .money")
		if len(price_discounted_arr) == 0 {
			price_discounted_arr = append(price_discounted_arr, "$0.00")
			is_unreleased = true
		}
		price_discounted := price_discounted_arr[0]

		if price_discounted == "" {
			price_discounted = price_original
		}

		iem_list = append(iem_list, Iem{title, price_original, price_discounted, is_unreleased, ""})
	})

	clt_shop.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		url_split := strings.SplitAfter(e.Request.URL.String(), "/")
		base_url := url_split[0] + url_split[1] + url_split[2]
		next_url := base_url + e.ChildAttr("a", "href")
		e.Request.Visit(next_url)
	})

	clt_shop.Visit(url_shop)
	clt_ranking.Visit(url_ranking)

	appendRankings(ranking_list, iem_list)
	app.clearCache()
	_ = app.createData("iems", iem_list, helpers.GetBaseUrl(url_shop))
	_ = app.createData("rankings", ranking_list, helpers.GetBaseUrl(url_ranking))
}
