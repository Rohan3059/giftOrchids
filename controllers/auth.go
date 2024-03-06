package controllers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohan3059/bizGrowth/database"
	"github.com/rohan3059/bizGrowth/models"
	generate "github.com/rohan3059/bizGrowth/tokens"
	"github.com/rohan3059/bizGrowth/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection = database.ProductData(database.Client, "User")

func SetOtpHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		contactNo := c.PostForm("mobileno")
		if contactNo == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "phone number can't be empty"})
			c.Abort()
			return
		}
		isNewUser := false
		filter := primitive.M{"mobileno": contactNo}
		res := UserCollection.FindOne(ctx, filter)
		err := res.Err()
		if err != nil && err != mongo.ErrNoDocuments {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}
		if err == mongo.ErrNoDocuments {
			user := models.USer{
				User_id:  primitive.NewObjectID(),
				MobileNo: contactNo,
			}
			UserCollection.InsertOne(ctx, user)
			isNewUser = true
		}
		otp, errG := generateOTP(contactNo)
		if errG != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
			c.Abort()
			return
		}
		update := primitive.M{
			"$set": primitive.M{
				"otp": otp,
			},
		}
		UserCollection.UpdateOne(ctx, filter, update)

		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "OTP sent successfully", "newUser": isNewUser})

	}
}
func ValidateOtpHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		contactNo := c.PostForm("mobileno")
		if contactNo == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "phone number can't be empty"})
			c.Abort()
			return
		}
		otp := c.PostForm("otp")
		if otp == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "otp can't be empty"})
			c.Abort()
			return
		}
		filter := primitive.M{utils.Mobileno: contactNo}
		res := UserCollection.FindOne(ctx, filter)
		err := res.Err()
		if err != nil && err != mongo.ErrNoDocuments {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}
		if err == mongo.ErrNoDocuments {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "mobile number doesn't exsist"})
			c.Abort()
			return
		}
		userDetails := models.USer{}
		dbErr := UserCollection.FindOne(ctx, filter).Decode(&userDetails)
		if dbErr != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong while decode"})
			c.Abort()
			return
		}

		if userDetails.OTP == otp {
			token, refreshtoken, _ := generate.UserTokenGenerator(userDetails.MobileNo, string(userDetails.User_id.Hex()))
			userDetails.Token = token
			userDetails.Refresh_token = refreshtoken
			update := primitive.M{
				"$set": primitive.M{
					"token":         token,
					"refresh_token": refreshtoken,
					"otp":           "",
				},
			}
			UserCollection.FindOneAndUpdate(ctx, filter, update)
			c.Header("content-type", "application/json")
			c.JSON(http.StatusAccepted, gin.H{"token": userDetails.Token})
		} else {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid OTP"})
		}
	}
}

func generateOTP(mobileNo string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	otp := 100000 + rand.Intn(900000)

	baseURL := "https://1.rapidsms.co.in/api/push"
	apiKey := "65be127aebd18"
	route := "TRANS"
	sender := "GRWTHB"
	message := `Welcome to Growth Biz! Your One-Time Password (OTP) for verification is ` + strconv.Itoa(otp) + ` Please use this code to complete the verification process. Do not share this code with anyone for security reasons."
	Visit:Growthbiz.co or mail info@growthbiz.co`
	// Encode the message
	encodedMessage := url.QueryEscape(message)

	// Construct the final URL
	finalURL := fmt.Sprintf("%s?apikey=%s&route=%s&sender=%s&mobileno=%s&text=%s", baseURL, apiKey, route, sender, mobileNo, encodedMessage)

	response, err := http.Get(finalURL)
	fmt.Println(response.Request.URL)
	//response.Request()
	if err != nil || response.StatusCode != 200 {
		fmt.Println("Error making GET request:", err)
		return "", errors.New("error ! plese try again")
	}
	return fmt.Sprintf("%06d", otp), nil
}
