package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostFeedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		var errors []string

		var feed models.Feed
		
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		 title := c.PostForm("title")
		 content :=c.PostForm("content")

		defer cancel()
		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]
		var uploadedURLs []string
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				log.Fatal(err)
				log.Println("error while opening file")
				c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
				return
			}
			uploadedURL, err := saveFile(f, file)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error saving file %s: %s", file.Filename, err.Error()))
			} else {
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
			if len(errors) > 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errors})
			} else {
				c.JSON(http.StatusOK, gin.H{"url": uploadedURLs})
			}
		}

		feed.FeedID = primitive.NewObjectID()
		feed.FeedDocument = uploadedURLs
		feed.Content = content
		feed.Title = title
		_, anyerr := FeedsCollection.InsertOne(ctx, feed)
		if anyerr != nil {
			fmt.Print(anyerr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Not Created"})
			return
		}

		c.JSON(http.StatusOK, "Successfully added feed!!")
		defer cancel()
	}
}
func GetAllFeedsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := FeedsCollection.Find(ctx,bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//iterate through cursor and get all feeds
		var feeds []models.Feed
		
		if err:= cursor.All(ctx,&feeds); err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"Error in cursor ": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, feeds)

	
	}

}

func DeleteFeed() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		feedId := c.Query("id")
		if feedId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "feedID can't be empty"})
			return
		}
		objID, err := primitive.ObjectIDFromHex(feedId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid feedID"})
			return
		}
		filter := primitive.M{"_id": objID}
		result, err := FeedsCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
			return
		}
		if result.DeletedCount < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to feed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}
