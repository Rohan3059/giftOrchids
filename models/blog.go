package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Blog struct {
	BlogID     primitive.ObjectID `json:"_id" bson:"_id"`
	Title      string             `json:"title" validate:"required"`
	SubTitle   string             `json:"subtitle" bson:"subtitle"`
	Slug       string             `json:"slug" bson:"slug"`
	Author     string             `json:"author" bson:"author"`
	ContentUrl string             `json:"contentUrl" bson:"contentUrl" validate:"required"`
	Keywords   []string           `json:"keywords" bson:"keywords"`
	Published  bool               `json:"published" bson:"published"`
	CoverImage string             `json:"coverImage" bson:"coverImage"`
	Created_at time.Time          `json:"created_at" bson:"created_at" `
	Updated_at time.Time          `json:"updated_at" bson:"updated_at" `
	IsArchived bool               `json:"isArchived" bson:"isArchived" `
}
