package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var contentCollection *mongo.Collection = database.ProductData(database.Client, "ContentItem")

func CreateContentItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var contentItem models.ContentItem
		defer cancel()

		ContentKey := c.PostForm("contentKey")
		if ContentKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Content can not be empty"})
			return
		}

		var existingContentItem models.ContentItem
		err := contentCollection.FindOne(ctx, bson.M{"contentKey": ContentKey}).Decode(&existingContentItem)
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while checking for existing content item"})
			return
		}
		if existingContentItem.ID != primitive.NilObjectID {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Content item with this key already exists"})
			return
		}

		Type := c.PostForm("type")
		if Type == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Content type can not be empty"})
			return
		}
		Description := c.PostForm("description")

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while parsing files"})
			return

		}

		var content interface{}
		if Type == "file" {
			files := form.File["content"]
			if files == nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Content file can not be empty"})
				return
			}

			var uploadedURLs []string
			for _, file := range files {
				f, err := file.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to open file: " + err.Error()})
					return
				}
				defer f.Close()

				uploadedURL, err := saveFile(f, file)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to save file: " + err.Error()})
					return
				}
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
			content = uploadedURLs

		} else {
			content = c.PostForm("content")
		}

		metadata := make(map[string]string)
		for key, values := range c.Request.PostForm {
			if key != "contentKey" && key != "type" && key != "description" && key != "content" {
				metadata[key] = values[0]
			}
		}

		contentItem = models.ContentItem{
			ID:          primitive.NewObjectID(),
			ContentKey:  ContentKey,
			Type:        Type,
			Description: Description,
			Content:     content,
			Metadata:    metadata,
			IsActive:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		result, err := contentCollection.InsertOne(ctx, contentItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"Status": http.StatusCreated, "Message": "success", "data": result})
	}
}

func GetContentItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		contentItemId := c.Param("contentItemId")
		var contentItem models.ContentItem
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(contentItemId)

		err := contentCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&contentItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": contentItem})
	}
}

func UpdateContentItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		contentItemId := c.Param("contentItemId")
		var contentItem models.ContentItem
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(contentItemId)

		err := contentCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&contentItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": err.Error()})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while parsing form"})
			return
		}

		var content interface{}
		if contentItem.Type == "file" {
			files := form.File["content"]
			if files == nil {
				c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "No image files provided"})
				return
			}

			var uploadedURLs []string
			for _, file := range files {
				f, err := file.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to open file: " + err.Error()})
					return
				}
				defer f.Close()

				uploadedURL, err := saveFile(f, file)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to save file: " + err.Error()})
					return
				}
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
			content = uploadedURLs
		} else {
			content = c.PostForm("content")
		}

		contentItem = models.ContentItem{
			Content:   content,
			UpdatedAt: time.Now(),
		}

		update := bson.M{
			"contentKey":  contentItem.ContentKey,
			"type":        contentItem.Type,
			"description": contentItem.Description,
			"content":     contentItem.Content,
			"metadata":    contentItem.Metadata,
			"is_active":   contentItem.IsActive,
			"updated_at":  time.Now(),
		}

		result, err := contentCollection.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		var updatedContentItem models.ContentItem
		if result.MatchedCount == 1 {
			err := contentCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&updatedContentItem)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": updatedContentItem})
	}
}

func DeleteContentItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		contentItemId := c.Param("id")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(contentItemId)

		result, err := contentCollection.DeleteOne(ctx, bson.M{"_id": objId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "Error": "Content item with specified ID not found!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Content item successfully deleted!"})
	}
}

func GetAllContentItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var contentItems []models.ContentItem
		defer cancel()

		results, err := contentCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleContentItem models.ContentItem
			if err = results.Decode(&singleContentItem); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			}
			contentItems = append(contentItems, singleContentItem)
		}
		for i, item := range contentItems {
			if content, ok := item.Content.(primitive.A); ok {
				updatedContent := make([]interface{}, len(content))
				for j, contentItem := range content {
					if contentItem != nil {
						if strItem, ok := contentItem.(string); ok {
							url, err := getPresignURL(strItem)
							if err == nil {
								updatedContent[j] = url
							} else {
								updatedContent[j] = ""
							}
						} else {
							updatedContent[j] = contentItem
						}
					}
				}
				contentItems[i].Content = updatedContent // Assign the updated slice back to Content
			}
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": contentItems})
	}
}

func GetContentItemsByKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		contentType := c.Param("contentKey")

		var contentItems models.ContentItem
		defer cancel()

		err := contentCollection.FindOne(ctx, bson.M{"contentKey": contentType}).Decode(&contentItems)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		if content, ok := contentItems.Content.(primitive.A); ok {
			updatedContent := make([]string, len(content))
			for i, item := range content {
				if item != nil {
					if strItem, ok := item.(string); ok {
						url, err := getPresignURL(strItem)
						if err == nil {
							updatedContent[i] = url
						} else {
							updatedContent[i] = ""
						}
					}
				}
			}
			contentItems.Content = updatedContent // Assign the updated slice back to Content
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": contentItems})
	}
}

func ToggleContentItemStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		contentItemId := c.Param("id")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(contentItemId)

		var contentItem models.ContentItem
		err := contentCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&contentItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": err.Error()})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"is_active":  !contentItem.IsActive,
				"updated_at": time.Now(),
			},
		}

		_, err = contentCollection.UpdateOne(ctx, bson.M{"_id": objId}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Content item status toggled successfully"})
	}
}

func UpdateFileContentItemContent() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		contentKey := c.Param("contentKey")
		if contentKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "Content key is required"})
			return
		}

		var contentItem models.ContentItem
		err := contentCollection.FindOne(ctx, bson.M{"contentKey": contentKey}).Decode(&contentItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Content item not found"})
			return
		}

		if contentItem.Type != "file" {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "This operation is only supported for file type content"})
			return
		}
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "Failed to parse multipart form"})
			return
		}

		files := form.File["content"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": "No files uploaded"})
			return
		}

		var uploadedURLs []string
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to open file"})
				return
			}
			defer src.Close()

			// Upload the file to S3 and get the URL
			uploadedURL, err := saveFile(src, file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to upload file"})
				return
			}
			uploadedURLs = append(uploadedURLs, uploadedURL)
		}

		var existingContent []string
		if contentItem.Content != nil {
			// Check if Content is of type primitive.A
			if contentArray, ok := contentItem.Content.(primitive.A); ok {
				// Convert primitive.A to []string
				for _, item := range contentArray {
					if str, ok := item.(string); ok {
						existingContent = append(existingContent, str)
					}
				}
			}
		}
		updatedContent := append(existingContent, uploadedURLs...)

		update := bson.M{
			"$set": bson.M{
				"content":    updatedContent,
				"updated_at": time.Now(),
			},
		}

		var updatedContentItem models.ContentItem
		err = contentCollection.FindOneAndUpdate(ctx, bson.M{"contentKey": contentKey}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedContentItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to fetch updated content item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": updatedContentItem})
	}
}

func DeleteImageFromContentItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		contentItemId := c.Param("contentItemId")
		indexStr := c.Param("index")

		objId, err := primitive.ObjectIDFromHex(contentItemId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Content Item ID"})
			return
		}

		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid index"})
			return
		}

		var contentItem models.ContentItem
		err = contentCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&contentItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Content Item not found"})
			return
		}

		// Check if Content is of type primitive.A
		var contentSlice []string
		if contentItem.Content != nil {
			if contentArray, ok := contentItem.Content.(primitive.A); ok {
				// Convert primitive.A to []string
				for _, item := range contentArray {
					if str, ok := item.(string); ok {
						contentSlice = append(contentSlice, str)
					}
				}
			}
		}

		// Check if the index is valid
		if index >= len(contentSlice) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Index out of range"})
			return
		}

		// Remove the image URL at the specified index
		contentSlice = append(contentSlice[:index], contentSlice[index+1:]...)

		// Update the Content field in the database
		update := bson.M{"$set": bson.M{"content": contentSlice}}
		_, err = contentCollection.UpdateOne(ctx, bson.M{"_id": objId}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update Content Item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "Image deleted successfully", "data": contentSlice})
	}
}
