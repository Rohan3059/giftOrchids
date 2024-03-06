package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohan3059/bizGrowth/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EnquiryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		mobileno, existse := c.Get("mobile")
		
		uid, exists := c.Get("uid")
		fmt.Print(mobileno)
		fmt.Print(uid)
		fmt.Print(exists)
		fmt.Print("Checking")
		fmt.Print(existse)
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
