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
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	generate "github.com/kravi0/BizGrowth-backend/tokens"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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

		created_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		if err == mongo.ErrNoDocuments {
			user := models.USer{
				User_id:    primitive.NewObjectID(),
				MobileNo:   contactNo,
				Created_at: created_at,
				Updated_at: created_at,
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

// RegisterUser handles the registration of a new user
func UpdateUserDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse mobile number from request
		mobileNo := c.PostForm("mobileno")
		if mobileNo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mobile number is not provided"})
			return
		}

		// Check if the user with the given mobile number exists
		var existingUser models.USer
		filter := bson.M{"mobileno": mobileNo}
		err := UserCollection.FindOne(context.Background(), filter).Decode(&existingUser)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Parse other user details from form data
		existingUser.UserName = c.PostForm("user_name")
		existingUser.Email = c.PostForm("email")
		existingUser.User_Address = c.PostForm("user_address")
		// Parse other fields from form data as needed

		// Save updated user details to the database
		update := bson.M{"$set": bson.M{
			"user_name":    existingUser.UserName,
			"email":        existingUser.Email,
			"user_address": existingUser.User_Address,
			// Add other fields as needed
		}}
		_, err = UserCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user details"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user details updated successfully"})
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

func ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			MobileNo string `json:"mobileno" validate:"required"`
			OTP      string `json:"otp" validate:"required"`
			Password string `json:"password" validate:"required,min=6"`
		}

		// Bind JSON request body to input struct
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		var seller models.Seller

		err := SellerCollection.FindOne(context.Background(), bson.M{"mobileno": input.MobileNo}).Decode(&seller)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "No Account is registered with this number"})
			return
		}

		if input.OTP != seller.OTP {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid OTP"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		filter := bson.M{"mobileno": input.MobileNo}
		update := bson.M{"$set": bson.M{"password": string(hashedPassword)}}
		_, err = SellerCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})

	}
}

func LoadUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exists := c.Get("uid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found in context"})
			return
		}

		// Convert sellerID to ObjectID
		userObjID, err := primitive.ObjectIDFromHex(userId.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Seller ID"})
			return
		}

		// Query the database to get seller information
		var user models.USer // Assuming Seller struct is defined in models package
		err = UserCollection.FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func RegisterAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name     string `json:"name" binding:"required"`
			Mobile   string `json:"mobile" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Check if the email or mobile number is already registered
		count, err := SellerCollection.CountDocuments(ctx, bson.M{"$or": []bson.M{
			{"email": input.Email},
			{"mobile": input.Mobile},
		}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Email or mobile number already registered"})
			return
		}

		// Hash the password before saving it to the database
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Generate a unique seller ID
		id := primitive.NewObjectID()
		token, refreshtoken, _ := generate.TokenGenerator(input.Email, input.Mobile, input.Name, id.Hex())

		admin := models.Seller{
			ID:            id,
			Seller_ID:     id.Hex(), // You can generate a unique seller ID here if needed
			Company_Name:  input.Name,
			MobileNo:      input.Mobile,
			Email:         input.Email,
			Password:      string(hashedPassword),
			Token:         token,
			Refresh_token: refreshtoken,
			User_type:     "ADMIN",
			Created_at:    time.Now(),
			Updated_at:    time.Now(),
		}

		// Insert the admin into the database
		_, err = SellerCollection.InsertOne(ctx, admin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Admin registered successfully"})
	}
}
