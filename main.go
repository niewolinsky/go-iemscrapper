package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Iem struct {
	Name                  string `bson:"name"`
	Price                 string `bson:"price"`
	Price_before_discount string `bson:"price_before_discount"`
	Is_unreleased         bool   `bson:"is_unreleased"`
}

type Ranking struct {
	Name  string `bson:"name"`
	Value string `bson:"value"`
	Rank  string `bson:"rank"`
}

type Metadata struct {
	Date      time.Time `bson:"date"`
	Start_url string    `bson:"start_url"`
	Scrap_id  int       `bson:"id"`
}

type ScrapResult struct {
	ScrappedData any
	Metadata     Metadata
}

type config struct {
	db struct {
		uri string
	}
	website struct {
		url string
	}
}

func initScrapper(url string) *colly.Collector {
	url_split := strings.SplitAfter(url, "/")
	base_url := url_split[1] + url_split[2]
	base_url_trim := base_url[1 : len(base_url)-1]

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

func dbConnect(cfg config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db_connection, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.db.uri))
	if err != nil {
		return nil, err
	}

	err = db_connection.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return db_connection, nil
}

func main() {
	cfg := config{}
	{
		cfg.website.url = "https://hifigo.com/collections/in-ear?sort_by=price-ascending"
		flag.StringVar(&cfg.db.uri, "db_uri", "", "MongoDB URI")
		flag.Parse()
	}

	db_connection, err := dbConnect(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to database failed: %v \n", err)
		os.Exit(1)
	}
	iems_collection := db_connection.Database("go-iemscrapper").Collection("iems")
	rank_crinnacle_collection := db_connection.Database("go-iemscrapper").Collection("rank_crinnacle")

	iem_list := []Iem{}
	ranking_list := []Ranking{}
	clt := initScrapper(cfg.website.url)
	crin := initScrapper("https://crinacle.com/")

	crin.OnHTML("#tablepress-4", func(e *colly.HTMLElement) {
		name := e.ChildTexts(".column-3")
		value := e.ChildTexts(".column-2")
		rank := e.ChildTexts(".column-1")

		for i := range name {
			ranking_list = append(ranking_list, Ranking{name[i], value[i], rank[i]})
		}
	})

	//scrap data from current page
	clt.OnHTML(".productitem--info", func(e *colly.HTMLElement) {
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

		fmt.Println("Im appending!")
		iem_list = append(iem_list, Iem{title, price_original, price_discounted, is_unreleased})
	})

	//visit next page
	clt.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		url_split := strings.SplitAfter(e.Request.URL.String(), "/")
		base_url := url_split[0] + url_split[1] + url_split[2]
		next_url := base_url + e.ChildAttr("a", "href")
		e.Request.Visit(next_url)
	})
	clt.Visit(cfg.website.url)
	crin.Visit("https://crinacle.com/rankings/iems/")

	result_iems := ScrapResult{iem_list, Metadata{time.Now(), cfg.website.url, 1}}
	result_rank_crinnacle := ScrapResult{ranking_list, Metadata{time.Now(), cfg.website.url, 1}}

	response, err := iems_collection.InsertOne(context.TODO(), result_iems)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.InsertedID)

	response, err = rank_crinnacle_collection.InsertOne(context.TODO(), result_rank_crinnacle)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.InsertedID)
}
