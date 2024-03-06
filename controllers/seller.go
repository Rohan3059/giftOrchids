package controllers

import (
	"context"
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
)

var SellerCollection *mongo.Collection = database.ProductData(database.Client, "seller")

// get all seller if no id is passes if id id passed it will return sepcific seller
func GetSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		if checkSeller(ctx, c) {
			var sellerDetail models.Seller
			uid, _ := c.Get("uid")
			uids := fmt.Sprintf("%v", uid)
			sellerID, _ := primitive.ObjectIDFromHex(string(uids))
			filter := primitive.M{"_id": sellerID}
			SellerCollection.FindOne(ctx, filter).Decode(&sellerDetail)
			c.JSON(http.StatusOK, sellerDetail)
			return
		}
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		var sellerDetails []models.Seller

		filter := bson.D{}
		Seller_ID := c.Query("sellerId")
		sellerID, _ := primitive.ObjectIDFromHex(Seller_ID)
		if sellerID.Hex() != "" {
			filter = bson.D{primitive.E{Key: "_id", Value: sellerID}}

		}
		cur, err := SellerCollection.Find(ctx, filter)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "can't find the user"})
			return
		}
		if err = cur.All(ctx, sellerDetails); err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		c.JSON(http.StatusOK, sellerDetails)
	}

}

// func AddSeller() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 		var sellerDetails models.Seller
// 		defer cancel()
// 	}
// }

// delete specific seller
func DeleteSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		seller_ID := c.Query("sellerID")
		if seller_ID == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid seller id"})
			c.Abort()
			return
		}
		sellerID, err := primitive.ObjectIDFromHex(seller_ID)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
		}
		filter := bson.D{primitive.E{Key: "_id", Value: sellerID}}
		_, err = SellerCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "fail to delete"})
			c.Abort()
			return
		}
		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "deleted successfully"})

	}
}
