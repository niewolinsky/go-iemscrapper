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

type UnmarshalHere struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Scrappeddata []Iem              `json:"scrappeddata" bson:"scrappeddata"`
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

func (app *application) readDataFromCache(key string) string {
	val2, err := app.cache_connection.Get(context.TODO(), key).Result()
	if err != nil {
		dataToCache := app.readDataFromDB("iems")
		err = app.cache_connection.Set(context.TODO(), key, dataToCache, time.Hour*24).Err()
		if err != nil {
			panic(err)
		}
		return dataToCache
	} else {
		return val2
	}
}

func (app *application) readDataFromDB(collection string) string {
	pick_collection := app.db_connection.Database("go-iemscrapper").Collection(collection)

	result := UnmarshalHere{}
	opts := options.FindOne().SetSort(bson.M{"$natural": -1})
	if err := pick_collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&result); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v \n", result)
	res, _ := json.MarshalIndent(result, "", "  ")
	return string(res)
}
