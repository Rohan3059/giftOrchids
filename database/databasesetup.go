package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {
		username := strings.TrimSpace("happyclarke")
    password := strings.TrimSpace("Ravi2580")

    uri := "mongodb+srv://" + username + ":" + password + "@my-cluster.p7z1vax.mongodb.net/?retryWrites=true&w=majority&appName=my-cluster"


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
	
	var collection = client.Database("growthbiz").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection = client.Database("growthbiz").Collection(collectionName)
	return productCollection
}

//func CategoriesData(client *mongo.Client, collectionNAme)
