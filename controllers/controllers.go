package controllers

import "go.mongodb.org/mongo-driver/mongo"

func NewApplication(prodCollection *mongo.Collection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}
