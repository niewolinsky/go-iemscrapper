package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/go-co-op/gocron"
)

type config struct {
	db struct {
		uri string
	}
	website struct {
		url string
	}
}

type application struct {
	config        config
	db_connection *mongo.Client
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

	app := &application{
		config:        cfg,
		db_connection: db_connection,
	}

	cron_job_scrapper := gocron.NewScheduler(time.UTC)
	cron_job_scrapper.Every(30).Seconds().Do(app.scrapData)
	cron_job_scrapper.StartAsync()

	app.serve()
}
