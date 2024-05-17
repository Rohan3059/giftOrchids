package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// get analytics related to seller registered, user, enquiry, from last 30 days and last 1 days
func GetAnalytics() gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {

			c.JSON(http.StatusForbidden, gin.H{"Error": "You're not authorized to access it"})
			return

		}

		days := c.Query("days")

		daysInt, _ := strconv.ParseInt(days, 10, 64)

		total_user_count, err := CountDocument(UserCollection, ctx, -1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		user_count, err := CountDocument(UserCollection, ctx, int(daysInt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		user_one_day_count, err := CountDocument(UserCollection, ctx, 1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		seller_total_count, err := CountDocument(SellerCollection, ctx, -1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		sellerCount, err := CountDocument(SellerCollection, ctx, int(daysInt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//one day seller count

		seller_one_day_count, err := CountDocument(SellerCollection, ctx, 1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		//total enquiry count
		total_enquiry_count, err := CountDocument(EnquireCollection, ctx, -1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		enquiry_count, err := CountDocument(EnquireCollection, ctx, int(daysInt))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		enquiry_one_day_count, err := CountDocument(EnquireCollection, ctx, 1)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		response := gin.H{

			"users": gin.H{
				"count":         user_count,
				"one_day_count": user_one_day_count,
				"total":         total_user_count,
			},
			"sellers": gin.H{
				"count":         sellerCount,
				"one_day_count": seller_one_day_count,
				"total":         seller_total_count,
			},
			"enquiries": gin.H{
				"count":         enquiry_count,
				"one_day_count": enquiry_one_day_count,
				"total":         total_enquiry_count,
			},
		}

		c.JSON(http.StatusOK, response)

	}
}

func CountDocument(collection *mongo.Collection, ctx context.Context, days int) (int64, error) {

	var filter primitive.M

	if days > 0 {
		pastTime := time.Now().AddDate(0, 0, -days)
		filter = bson.M{
			"created_at": bson.M{
				"$gte": primitive.NewDateTimeFromTime(pastTime)},
		}
	} else {
		filter = bson.M{}
	}

	count, err := collection.CountDocuments(ctx, filter)
	return count, err
}
