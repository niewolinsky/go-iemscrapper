package main

import (
	"context"
	"fmt"
	"time"
)

type Metadata struct {
	Date      time.Time `bson:"date"`
	Start_url string    `bson:"start_url"`
	Scrap_id  int       `bson:"id"`
}

type ScrapResult struct {
	ScrappedData any
	Metadata     Metadata
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

func (app *application) readData() {
	fmt.Println("asd")
}
