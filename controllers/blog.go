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

// PublishBlog updates a blog post to set it as published
func PublishBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := c.Param("id")
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid blog ID"})
			return
		}

		// Update the blog to set Published to true and IsArchived to false
		update := bson.M{
			"$set": bson.M{
				"published":   true,
				"is_archived": false,
			},
		}

		result, err := BlogCollection.UpdateOne(ctx, bson.M{"_id": objId}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update blog"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Blog not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "Blog published successfully"})
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
		ctx := context.Background()

		projection := bson.D{
			{Key: "_id", Value: 1},
			{Key: "title", Value: 1},
			{Key: "coverImage", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "keywords", Value: 1},
			{Key: "author", Value: 1},
			{Key: "published", Value: 1},
		}

		cursor, err := BlogCollection.Find(ctx, bson.D{}, options.Find().SetProjection(projection))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blogs"})
			return
		}
		defer cursor.Close(ctx)

		var blogs []struct {
			BlogID     primitive.ObjectID `bson:"_id" json:"id"`
			Title      string             `bson:"title" json:"title"`
			CoverImage string             `bson:"coverImage" json:"coverImage"`
			CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
			Author     string             `bson:"author" json:"author"`
			Keywords   []string           `bson:"keywords" json:"keywords"`
			Published  bool               `bson:"published" json:"published"`
		}

		if err := cursor.All(ctx, &blogs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode blogs"})
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

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": blogs})
	}
}

// get blog by id
func GetBlogByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
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
// UpdateBlog updates an existing blog post
func UpdateBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			return
		}

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

		var url string

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
				url = uploadedURL
			}

		}
		// Prepare the update document with the $set operator
		setUpdate := bson.M{}

		if title != "" {
			setUpdate["title"] = title
		}
		if slug != "" {
			setUpdate["slug"] = slug
		}
		if subtitle != "" {
			setUpdate["subtitle"] = subtitle
		}
		if content != "" {
			setUpdate["contentUrl"] = content
		}
		if author != "" {
			setUpdate["author"] = author
		}
		if keywords != "" {
			setUpdate["keywords"] = keywords
		}
		if url != "" {
			setUpdate["coverImage"] = url
		}

		// Now create the update document
		update := bson.M{"$set": setUpdate}
		// Perform the update
		result := BlogCollection.FindOneAndUpdate(context.Background(), bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update blog"})
			return
		}

		if result.Err() != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": result.Err().Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "Blog updated successfully", "data": result})
	}
}

// delete blog
func DeleteBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You're not authroize for this."})
			return
		}

		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "Invalid ID"})
			return
		}

		result, err := BlogCollection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Blog not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Blog deleted successfully"})
	}
}

// GetBlogBySlug retrieves a blog post by its slug
func GetBlogBySlug() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		slug := c.Param("slug")
		var blog models.Blog

		// Find the blog by slug
		err := BlogCollection.FindOne(ctx, bson.M{"slug": slug, "published": true}).Decode(&blog)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Blog not found"})
			return
		}

		blog.CoverImage, err = getPresignURL(blog.CoverImage)
		if err != nil {
			blog.CoverImage = ""
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": blog})
	}
}

// GetPublishedBlogs retrieves all published blog posts
func GetPublishedBlogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := BlogCollection.Find(ctx, bson.M{"published": true}) // Filter for published blogs
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var blogs []models.Blog
		if err = cursor.All(ctx, &blogs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}
		fmt.Print(blogs)
		if blogs == nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNoContent, "Message": "error", "data": "No blog found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": blogs})
	}
}

// ArchiveBlog archives a blog post
func ArchiveBlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You're not authroize for this."})
			return
		}

		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "Invalid ID"})
			return
		}

		update := bson.M{"$set": bson.M{"isArchived": true, "published": "false"}} // Set IsArchived to true
		result, err := BlogCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Blog not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Blog archived successfully"})
	}
}
