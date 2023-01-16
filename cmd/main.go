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
	"github.com/go-redis/redis/v9"
)

type config struct {
	db struct {
		uri string
	}
	cache struct {
		uri string
	}
	website struct {
		url string
	}
}

type application struct {
	config           config
	db_connection    *mongo.Client
	cache_connection *redis.Client
}

func cacheConnect(cfg config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.cache.uri,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
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
		flag.StringVar(&cfg.cache.uri, "cache_uri", "", "Redis URI")
		flag.Parse()
	}

	db_connection, err := dbConnect(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to database failed: %v \n", err)
		os.Exit(1)
	}
	defer func() {
		if err = db_connection.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	redis_connection, err := cacheConnect(cfg)
	defer redis_connection.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "connecting to cache (redis) failed: %v \n", err)
		os.Exit(1)
	}

	app := &application{
		config:           cfg,
		db_connection:    db_connection,
		cache_connection: redis_connection,
	}

	cron_job_scrapper := gocron.NewScheduler(time.UTC)
	cron_job_scrapper.Every(12).Hours().Do(app.scrapData)
	cron_job_scrapper.StartAsync()

	app.serve()
}
