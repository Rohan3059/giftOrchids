package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var AttributesCollection *mongo.Collection = database.ProductData(database.Client, "AttributeType")

func GetAllAttributes() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := AttributesCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		var attributes []bson.M
		if err = cursor.All(ctx, &attributes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, attributes)
	}
}

// @Summary Add new attribute type to the

func AddAttributeType() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var attribute models.AttributeType
		if err := c.BindJSON(&attribute); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		attribute.ID = primitive.NewObjectID()
		if attribute.Attribute_Name == "" || attribute.Attribute_Code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return

		}
		_, err := AttributesCollection.InsertOne(ctx, attribute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, attribute)
	}
}

//get attribute by id

func GetAttributeByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		var attribute models.AttributeType
		err = AttributesCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&attribute)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Attribute not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, attribute)
	}
}

//update attribute

func UpdateAttributeType() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			return
		}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			return
		}
		var attribute models.AttributeType
		err = AttributesCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&attribute)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Attribute not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		if err := c.BindJSON(&attribute); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err})
			return
		}
		attribute.ID = oid
		_, err = AttributesCollection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": attribute})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Attribute has been updated succesfully"})
	}
}
