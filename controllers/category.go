package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var CategoriesCollection *mongo.Collection = database.ProductData(database.Client, "Categories")

func AddCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		var category models.Categories
		defer cancel()
		if err := c.BindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		category.Category_ID = primitive.NewObjectID()
		count, err := CategoriesCollection.CountDocuments(ctx, bson.M{"category": category.Category})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category already exist"})
			return
		}
		category.Category_ID = primitive.NewObjectID()
		_, anyerr := CategoriesCollection.InsertOne(ctx, category)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Category Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Succesfully added category")

	}
}
func GetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {

		var category []models.Categories
		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cursor, err := CategoriesCollection.Find(ctx, bson.D{})
		if err != nil {

			c.JSON(http.StatusInternalServerError, "something went worng please try after some time")
			return
		}

		if err = cursor.All(ctx, &category); err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, "something went wrong while fetching data Try Again")
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		c.IndentedJSON(http.StatusOK, category)

	}
}

func EditCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cat_id := c.Query("cat_id")

		if cat_id == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid user id"})
			c.Abort()
			return
		}

		catID, err := primitive.ObjectIDFromHex(cat_id)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
			return
		}

		var Editcategory models.Categories
		if err := c.BindJSON(&Editcategory); err != nil {
			c.IndentedJSON(http.StatusBadRequest, "category is not in correct format")
			return
		}
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: catID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "category", Value: Editcategory.Category}}}}
		_, err = CategoriesCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully updated category")

	}
}
