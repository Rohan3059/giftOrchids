package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EnquiryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		mobileno, existse := c.Get("mobile")
		
		uid, exists := c.Get("uid")
	
		if !existse || !exists || uid == "" || mobileno == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		var enquire models.Enquire
		if err := c.BindJSON(&enquire); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

	

		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		enquire.Enquire_id = primitive.NewObjectID()
		
		
		_, err := EnquireCollection.InsertOne(ctx, enquire)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Success": "enquiry registerd"})
	}
}

func GETEnquiryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin Token Not found"})
			return
		}

		var enquire []models.Enquire

		cursor, err := EnquireCollection.Find(ctx, primitive.M{})
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch"})
		}
		if err := cursor.All(ctx, &enquire); err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		defer cursor.Close(context.Background())
		c.JSON(http.StatusOK, enquire)
	}
}

// GetAllRequirementMessages retrieves all RequirementMessages
func GetAllRequirementMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := RequirementMessageCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var messages []models.RequirementMessage
		if err := cursor.All(ctx, &messages); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}

// CreateRequirementMessage creates a new RequirementMessage
func CreateRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate new ObjectID
		message.Requirement_id = primitive.NewObjectID()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := RequirementMessageCollection.InsertOne(ctx, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, result.InsertedID)
	}
}
// UpdateRequirementMessage updates a RequirementMessage
func UpdateRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": message.Requirement_id}
		update := bson.M{"$set": message}

		_, err := RequirementMessageCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "RequirementMessage updated successfully"})
	}
}

// DeleteRequirementMessage deletes a RequirementMessage
func DeleteRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": message.Requirement_id}

		_, err := RequirementMessageCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "RequirementMessage deleted successfully"})
	}
}
