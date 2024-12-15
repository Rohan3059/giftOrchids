package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {

	/*if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}*/

	username := url.QueryEscape(os.Getenv("DB_USER"))
	password := url.QueryEscape(os.Getenv("DB_PASSWORD"))

	uri := "mongodb+srv://" + username + ":" + password + "@grwothbiz.srweepy.mongodb.net/?retryWrites=true&w=majority&appName=GrwothBiz"

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Print(err)
		log.Println("failed to connect to mnongodb")
		return nil
	}

	fmt.Println("connected to mongodb")
	return client

}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection = client.Database("giftOrchids").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection = client.Database("giftOrchids").Collection(collectionName)
	return productCollection
}

//func CategoriesData(client *mongo.Client, collectionNAme)
