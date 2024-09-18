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

var settingsCollection *mongo.Collection = database.ProductData(database.Client, "Settings")

func CreateSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var setting models.Settings
		if err := c.ShouldBindJSON(&setting); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": err.Error()})
			return
		}

		setting.ID = primitive.NewObjectID()
		setting.Created_at = time.Now()
		setting.Updated_at = time.Now()

		result, err := settingsCollection.InsertOne(ctx, setting)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to create setting"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"Status": http.StatusCreated, "Message": "success", "data": result})
	}
}

// GetSetting retrieves a setting by its name
func GetSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		name := c.Param("name")
		var setting models.Settings

		err := settingsCollection.FindOne(ctx, bson.M{"name": name}).Decode(&setting)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Setting not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": setting})
	}
}

func UpdateSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		name := c.Param("name")
		var setting models.Settings
		if err := c.ShouldBindJSON(&setting); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Status": http.StatusBadRequest, "Message": "error", "data": err.Error()})
			return
		}

		setting.Updated_at = time.Now()

		update := bson.M{
			"$set": bson.M{
				"value":      setting.Value,
				"updated_at": setting.Updated_at,
			},
		}

		result, err := settingsCollection.UpdateOne(ctx, bson.M{"name": name}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to update setting"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Setting not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Setting updated successfully"})

	}
}

func GetSettingByName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		name := c.Param("name")
		var setting models.Settings

		err := settingsCollection.FindOne(ctx, bson.M{"name": name}).Decode(&setting)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Setting not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to retrieve setting"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": setting})
	}
}

func DeleteSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		name := c.Param("name")

		result, err := settingsCollection.DeleteOne(ctx, bson.M{"name": name})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Status": http.StatusInternalServerError, "Message": "error", "data": "Failed to delete setting"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Status": http.StatusNotFound, "Message": "error", "data": "Setting not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Status": http.StatusOK, "Message": "success", "data": "Setting deleted successfully"})
	}
}
