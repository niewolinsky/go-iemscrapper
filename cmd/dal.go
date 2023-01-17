package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Metadata struct {
	Date      time.Time `json:"date" bson:"date"`
	Start_url string    `json:"start_url" bson:"start_url"`
	Scrap_id  int       `json:"id" bson:"id"`
}

type ScrapResult struct {
	ScrappedData any
	Metadata     Metadata
}

type ResponseTemplate struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Scrappeddata []Iem              `json:"scrappeddata" bson:"scrappeddata"`
	Metadata
}

type ResponseTemplate2 struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Scrappeddata []Ranking          `json:"scrappeddata" bson:"scrappeddata"`
	Metadata
}

func (app *application) createData(collection string, data any, url string) (inserted_id string) {
	pick_collection := app.db_connection.Database("go-iemscrapper").Collection(collection)
	result := ScrapResult{data, Metadata{time.Now(), url, 1}}

	response, err := pick_collection.InsertOne(context.TODO(), result)
	if err != nil {
		panic(err)
	}

	inserted_id = fmt.Sprintf("value: %v", response.InsertedID)
	return inserted_id
}

func (app *application) clearCache() {
	app.cache_connection.FlushAll(context.TODO())
}

func (app *application) getLatestScrapCache(key string) string {
	result, err := app.cache_connection.Get(context.TODO(), key).Result()
	if err != nil {
		dataToCache := app.getLatestScrap("iems")
		err = app.cache_connection.Set(context.TODO(), key, dataToCache, time.Hour*24).Err()
		if err != nil {
			panic(err)
		}
		return dataToCache
	} else {
		return result
	}
}

func (app *application) getLatestScrap(collection string) string {
	pick_collection := app.db_connection.Database("go-iemscrapper").Collection(collection)

	response_template := ResponseTemplate{}

	opts := options.FindOne().SetSort(bson.M{"$natural": -1})
	err := pick_collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&response_template)
	if err != nil {
		log.Fatal(err)
	}

	result, _ := json.MarshalIndent(response_template, "", "  ")
	return string(result)
}

func (app *application) getAllScraps(collection string) string {
	pick_collection := app.db_connection.Database("go-iemscrapper").Collection(collection)

	filterCursor, err := pick_collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	iemsFiltered := []ResponseTemplate{}
	err = filterCursor.All(context.TODO(), &iemsFiltered)
	if err != nil {
		log.Fatal(err)
	}

	result, _ := json.MarshalIndent(iemsFiltered, "", "  ")
	return string(result)
}
