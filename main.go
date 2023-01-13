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

type Metadata struct {
	Date      time.Time `bson:"date"`
	Start_url string    `bson:"start_url"`
	Scrap_id  int       `bson:"id"`
}

type ScrapResult struct {
	Iems     []Iem
	Metadata Metadata
}

type config struct {
	db struct {
		uri string
	}
	website struct {
		url string
	}
}

func initScrapper() *colly.Collector {
	clt := colly.NewCollector()

	clt.OnError(func(_ *colly.Response, err error) {
		fmt.Fprintf(os.Stderr, "error while scraping the website: %v \n", err)
		os.Exit(1)
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

	iem_list := []Iem{}
	clt := initScrapper()
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

		iem_list = append(iem_list, Iem{title, price_original, price_discounted, is_unreleased})
	})
	//visit next page
	clt.OnHTML(".pagination--next", func(e *colly.HTMLElement) {
		urlSplit := strings.SplitAfter(e.Request.URL.String(), "/")
		baseUrl := urlSplit[0] + urlSplit[1] + urlSplit[2]
		nextUrl := baseUrl + e.ChildAttr("a", "href")
		e.Request.Visit(nextUrl)
	})
	clt.Visit(cfg.website.url)

	result := ScrapResult{iem_list, Metadata{time.Now(), cfg.website.url, 1}}

	response, err := iems_collection.InsertOne(context.TODO(), result)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.InsertedID)
}
