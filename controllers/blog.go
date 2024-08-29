package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var BlogCollection *mongo.Collection = database.ProductData(database.Client, "Blog")

func CreateBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		var blog models.Blog

		title := c.PostForm("title")
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "title is required"})
			return
		}

		slug := c.PostForm("slug")

		if slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "slug is required"})
			return
		}

		subtitle := c.PostForm("subtitle")

		content := c.PostForm("blogContent")

		author := c.PostForm("author")

		keywords := c.PostForm("keywords")

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["cover"]

		if len(files) != 0 {

			f, err := files[0].Open()
			if err != nil {
				c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
				return
			}
			uploadedURL, err := saveFile(f, files[0])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			} else {
				blog.CoverImage = uploadedURL
			}

			blog.CoverImage = uploadedURL

		}

		blog.BlogID = primitive.NewObjectID()
		blog.Created_at = time.Now()
		blog.Updated_at = time.Now()
		blog.ContentUrl = content

		blog.Title = title
		blog.SubTitle = subtitle
		blog.Author = author
		blog.Slug = slug
		blog.Published = false
		blog.IsArchived = false

		if keywords != "" {
			//unmarshal keyowrds as array of string
			var keywordsArray []string
			err := json.Unmarshal([]byte(keywords), &keywordsArray)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}
			blog.Keywords = keywordsArray
		}

		_, insertErr := BlogCollection.InsertOne(context.Background(), blog)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to create blog post"})
			return
		}

		c.JSON(http.StatusCreated, blog)
	}

}

// publish blog
func PublishBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var blog models.Blog
		err := BlogCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&blog)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get blog"})
			return
		}
		blog.Published = true
		_, err = BlogCollection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": blog})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update blog"})
			return
		}
		c.JSON(http.StatusOK, blog)
	}
}

// get all blogs
func GetAllBlogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := BlogCollection.Find(context.Background(), bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get blogs"})
			return
		}
		var blogs []models.Blog
		err = cursor.All(context.Background(), &blogs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get blogs"})
			return
		}
		c.JSON(http.StatusOK, blogs)
	}
}

// get only title, cover, created_at for all blogs
func GetBlogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define the projection to select only the desired fields
		projection := bson.D{
			{Key: "_id", Value: 1},
			{Key: "title", Value: 1},
			{Key: "coverImage", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "author", Value: 1},
		}

		// Fetch the blogs with the defined projection
		cursor, err := BlogCollection.Find(context.Background(), bson.D{}, options.Find().SetProjection(projection))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blogs"})
			return
		}
		defer cursor.Close(context.Background())

		var blogs []struct {
			BlogID     primitive.ObjectID `bson:"_id"`
			Title      string             `json:"title"`
			CoverImage string             `json:"coverImage"`
			CreatedAt  time.Time          `json:"created_at"`
			Author     string             `json:"author"`
		}

		// Decode the cursor into the blogs slice
		err = cursor.All(context.Background(), &blogs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to parse blogs"})
			return
		}

		for i, blog := range blogs {
			if blog.CoverImage != "" {
				blogs[i].CoverImage, err = getPresignURL(blogs[i].CoverImage)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to generate presigned URL"})
					return
				}
			}
		}

		// Respond with the selected blog fields
		c.JSON(http.StatusOK, blogs)
	}
}

// get blog by id
func GetBlogByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fmt.Print("ID=" + id)
		//convrt to objectid
		objID, er := primitive.ObjectIDFromHex(id)
		if er != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			return
		}
		var blog models.Blog
		err := BlogCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&blog)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get blog"})
			return
		}

		blog.CoverImage, err = getPresignURL(blog.CoverImage)
		if err != nil {
			blog.CoverImage = ""
		}

		c.JSON(http.StatusOK, blog)
	}
}

// update blog
func UpdateBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var blog models.Blog
		if err := c.ShouldBindJSON(&blog); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		filter := bson.M{"_id": id}
		update := bson.M{"$set": blog}
		_, err := BlogCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update blog"})
			return
		}
		c.JSON(http.StatusOK, blog)
	}
}

// delete blog
func DeleteBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := BlogCollection.DeleteOne(context.Background(), bson.M{"_id": id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to delete blog"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Blog deleted successfully"})
	}
}
