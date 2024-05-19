package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// update user profile
func UpdateUserProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		uid, exist := c.Get("uid")

		if !exist {
			c.JSON(http.StatusForbidden, gin.H{"Error": "You're not authorized to access it"})
			return
		}

		var user models.USer

		userId, _ := primitive.ObjectIDFromHex(uid.(string))

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println(user)

		email := user.Email
		mobile := user.MobileNo
		name := user.UserName
		address := user.User_Address

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		filter := bson.M{"_id": userId}

		//create update otpions if these values are not empty
		update := []primitive.M{}
		if email != "" {
			update = append(update, bson.M{"$set": bson.M{"email": email}})
		}

		if mobile != "" {
			update = append(update, bson.M{"$set": bson.M{"mobileno": mobile}})
		}

		if name != "" {
			update = append(update, bson.M{"$set": bson.M{"user_name": name}})
		}

		if address != "" {
			update = append(update, bson.M{"$set": bson.M{"address": address}})
		}

		update = append(update, bson.M{"$set": bson.M{"updated_at": updated_at}})

		var foundUser models.USer

		updateErr := UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&foundUser)

		if updateErr != nil {
			if errors.Is(updateErr, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"Error": "user not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"Error": updateErr.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": "profile updated successfully",
		})

	}
}
