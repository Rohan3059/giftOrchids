package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddReviewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request body to get review details
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var review models.Reviews
		if err := c.ShouldBindJSON(&review); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userId, exist := c.Get("uid")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(userId.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		review.UserId = oid
		//not null userId
		if review.UserId.Hex() == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You're not logged in"})
			return
		}

		if review.ReviewsDetails.ReviewRating > 5 || review.ReviewsDetails.ReviewRating < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
			return
		}

		if review.ReviewsDetails.ReviewText == "" || review.ReviewsDetails.ReviewTitle == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review text and title are required"})
			return
		}

		// Check if product exists
		var product models.Product
		productObjID, err := primitive.ObjectIDFromHex(review.ProductId.Hex())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		filter := bson.M{"_id": productObjID}
		err = ProductCollection.FindOne(ctx, filter).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Check if user has already reviewed the product
		filter = bson.M{"product_id": productObjID, "user_id": review.UserId}
		count, err := ReviewsCollection.CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if user has already reviewed the product"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reviewed this product"})
			return
		}

		review.Id = primitive.NewObjectID()
		review.CreatedAt = time.Now()
		review.UpdatedAt = time.Now()

		_, errs := ReviewsCollection.InsertOne(ctx, review)
		if errs != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
			return
		}

		if product.Reviews != nil {

			update := bson.M{"$push": bson.M{"reviews": review.Id}}
			_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": review.ProductId}, update)
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with review ID"})
				return
			}
		} else {

			update := bson.M{"$set": bson.M{"reviews": []primitive.ObjectID{review.Id}}}
			_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": review.ProductId}, update)
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with review ID"})
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Thanks for your review!"})
	}
}

func AddReviewByAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var review models.Reviews
		if err := c.ShouldBindJSON(&review); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !checkAdmin(ctx, c) {

			c.JSON(http.StatusForbidden, gin.H{"Error": "You're not authorized to add reviews"})
			return
		}

		userId := c.Query("userId")

		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User ID is required"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		review.UserId = oid

		if review.ReviewsDetails.ReviewRating > 5 || review.ReviewsDetails.ReviewRating < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Rating must be between 1 and 5"})
			return
		}

		if review.ReviewsDetails.ReviewText == "" || review.ReviewsDetails.ReviewTitle == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Review text and title are required"})
			return
		}

		// Check if product exists
		var product models.Product
		pid := review.ProductId

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid product ID"})
			return
		}
		productfilter := bson.M{"_id": pid}
		err = ProductCollection.FindOne(ctx, productfilter).Decode(&product)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
			return
		}

		// Check if user has already reviewed the product
		filter := bson.M{"product_id": product.Product_ID, "user_id": review.UserId}
		count, err := ReviewsCollection.CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if user has already reviewed the product"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User have already reviewed this product"})
			return
		}

		review.Id = primitive.NewObjectID()
		review.CreatedAt = time.Now()
		review.UpdatedAt = time.Now()

		_, errs := ReviewsCollection.InsertOne(ctx, review)
		if errs != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
			return
		}

		if product.Reviews != nil {

			update := bson.M{"$push": bson.M{"reviews": review.Id}}
			_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": review.ProductId}, update)
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with review ID"})
				return
			}
		} else {

			update := bson.M{"$set": bson.M{"reviews": []primitive.ObjectID{review.Id}}}
			_, err = ProductCollection.UpdateOne(ctx, productfilter, update)
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with review ID"})
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Thanks for your review!"})
	}
}

func ApproveReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {

			c.JSON(http.StatusForbidden, gin.H{"Error": "You're not authorized to approve reviews"})
			return

		}

		status := c.Query("status")

		//parse status bool
		statusBool, _ := strconv.ParseBool(status)

		id := c.Param("id")

		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Review ID is required"})
			return
		}

		oid, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid review ID"})
			return
		}

		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"approved": statusBool}}
		_, updteErr := ReviewsCollection.UpdateOne(ctx, filter, update)
		if updteErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to approve review"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Review status updated successfully"})

	}
}

func GetProductApprovedReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var productId = c.Query("product_id")
		if productId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Product ID is required"})
			return
		}

		var product models.Product
		productObjID, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		filter := bson.M{"_id": productObjID}
		err = ProductCollection.FindOne(ctx, filter).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		pipeline := []bson.M{
			{
				"$match": bson.M{"product_id": productObjID, "approved": true},
			},
			{
				"$lookup": bson.M{
					"from":         "User",
					"localField":   "user_id",
					"foreignField": "_id",
					"as":           "user",
				},
			},
			{
				"$unwind": bson.M{
					"path":                       "$user",
					"preserveNullAndEmptyArrays": true,
				},
			},
			{
				"$project": bson.M{
					"_id":             1,
					"product_id":      1,
					"reviews_details": 1,
					"approved":        1,
					"archived":        1,
					"created_at":      1,

					"user": bson.M{
						"_id":      "$user._id",
						"name":     "$user.username",
						"mobileno": "$user.mobileno",
					},
				},
			},
		}

		//get reviews with product_id
		cursor, err := ReviewsCollection.Aggregate(ctx, pipeline)
		if err != nil {
			fmt.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to find reviews"})
			return
		}
		defer cursor.Close(ctx)

		var result []bson.M
		if err := cursor.All(ctx, &result); err != nil {
			fmt.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get reviews"})
			return
		}

		var totalRating int32 = 0
		totalReviews := len(result)
		ratingCounts := make(map[int32]int)

		// Iterate through each review in the result slice
		for _, review := range result {
			rating := review["reviews_details"].(bson.M)["review_rating"].(int32)

			// Calculate total rating
			totalRating += rating

			// Count the number of reviews for each rating
			ratingCounts[rating]++
		}

		// Calculate average rating
		averageRating := float64(totalRating) / float64(totalReviews)

		// Calculate percentage of reviews for each ReviewRating
		percentageReviews := make(map[int32]float64)
		for rating, count := range ratingCounts {
			percentage := float64(count) / float64(totalReviews) * 100
			percentageReviews[rating] = percentage
		}

		c.JSON(http.StatusOK, gin.H{
			"reviews":          result,
			"totalReviews":     totalReviews,
			"averageRating":    averageRating,
			"ratingPercentage": percentageReviews,
		})

	}
}

func GetProductReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var productId = c.Query("product_id")
		if productId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Product ID is required"})
			return
		}

		var product models.Product
		productObjID, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		filter := bson.M{"_id": productObjID}
		err = ProductCollection.FindOne(ctx, filter).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		pipeline := []bson.M{
			{
				"$match": bson.M{"product_id": productObjID},
			},
			{
				"$lookup": bson.M{
					"from":         "User",
					"localField":   "user_id",
					"foreignField": "_id",
					"as":           "user",
				},
			},
			{
				"$unwind": bson.M{
					"path":                       "$user",
					"preserveNullAndEmptyArrays": true,
				},
			},
			{
				"$project": bson.M{
					"_id":             1,
					"product_id":      1,
					"reviews_details": 1,
					"approved":        1,
					"archived":        1,
					"created_at":      1,

					"user": bson.M{
						"_id":      "$user._id",
						"name":     "$user.username",
						"mobileno": "$user.mobileno",
					},
				},
			},
		}

		//get reviews with product_id
		cursor, err := ReviewsCollection.Aggregate(ctx, pipeline)
		if err != nil {
			fmt.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to find reviews"})
			return
		}
		defer cursor.Close(ctx)

		var result []bson.M
		if err := cursor.All(ctx, &result); err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to get reviews"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func GetReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var reviews []bson.M

		pipeline := []bson.M{

			{
				"$lookup": bson.M{
					"from":         "User",
					"localField":   "user_id",
					"foreignField": "_id",
					"as":           "user",
				},
			},
			{
				"$lookup": bson.M{
					"from":         "Products",
					"localField":   "product_id",
					"foreignField": "_id",
					"as":           "product",
				},
			},
			{
				"$unwind": bson.M{
					"path":                       "$user",
					"preserveNullAndEmptyArrays": true,
				},
			}, {
				"$unwind": bson.M{
					"path":                       "$product",
					"preserveNullAndEmptyArrays": true,
				},
			},
			{
				"$project": bson.M{
					"_id":             1,
					"product":         1,
					"reviews_details": 1,
					"approved":        1,
					"archived":        1,
					"created_at":      1,

					"user": bson.M{
						"_id":      "$user._id",
						"name":     "$user.username",
						"mobileno": "$user.mobileno",
					},
				},
			},
		}
		cursor, err := ReviewsCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong while fetching the reviews")
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var review bson.M
			err := cursor.Decode(&review)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, "Something went wrong while fetching the reviews")
				return
			}
			reviews = append(reviews, review)
		}
		c.IndentedJSON(http.StatusOK, reviews)

	}
}

func GetReview() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := c.Param("id")

		objID, _ := primitive.ObjectIDFromHex(id)

		var reviews models.Reviews

		err := ReviewsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&reviews)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong while fetching the review")
			return
		}

		//get productid and find product details
		productid := reviews.ProductId

		prodcut := getProductDetails(ctx, productid.Hex())

		userid := reviews.UserId.Hex()

		user := getUserDetails(ctx, userid)

		c.IndentedJSON(http.StatusOK, gin.H{"review": reviews, "product": prodcut, "user": user})

	}
}
