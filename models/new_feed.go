package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Feed struct {
	FeedID       primitive.ObjectID `bson:"_id"`
	Title        string             `json:"title" validate:"required"`
	Content      string             `json:"content" validate:"required"`
	FeedDocument []string           `json:"feedDocument" validate:"required"`
}
