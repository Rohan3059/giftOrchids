package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	generate "github.com/kravi0/BizGrowth-backend/tokens"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)

}

var validate = validator.New()

var SellerTmpCollection *mongo.Collection = database.ProductData(database.Client, "SellerTmp")


/* seller registartion */


func SellerRegistrationSendOTP() gin.HandlerFunc {
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
		filter := primitive.M{"mobileno": contactNo}
		count, err := SellerTmpCollection.CountDocuments(ctx, filter)
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Phone is already in use"})
			return
		}
		Seller := models.SellerTmp{
			ID:       primitive.NewObjectID(),
			MobileNo: contactNo,
		}
		SellerTmpCollection.InsertOne(ctx, Seller)
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
		SellerTmpCollection.UpdateOne(ctx, filter, update)

		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "OTP sent successfully"})

	}
}
func SellerRegistrationOtpVerification() gin.HandlerFunc {
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
		res := SellerTmpCollection.FindOne(ctx, filter)
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
		SellerDetails := models.SellerTmp{}
		dbErr := SellerTmpCollection.FindOne(ctx, filter).Decode(&SellerDetails)
		if dbErr != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong while decode"})
			c.Abort()
			return
		}

		if SellerDetails.OTP == otp {
			SellerTmpCollection.FindOneAndUpdate(ctx, filter, primitive.M{"otp": ""})
			c.Header("content-type", "application/json")
			c.JSON(http.StatusOK, gin.H{"success": "verified"})
		} else {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid OTP"})
		}
	}
}

func SellerRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var seller models.Seller
		mobileno := c.PostForm("mobileno")
		Company_Name := c.PostForm("Company_name")
		NameOfOwner := c.PostForm("nameofowner")
		AadharNumber := c.PostForm("aadharnumber")
		PAN := c.PostForm("pan")
		PermanentAddress := c.PostForm("permanenetaddress")
		email := c.PostForm("email")
		password := HashPassword(c.PostForm("password"))
	
	

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		panFile := form.File["panFile"]
		aadharFile := form.File["aadharFile"]


		panHeader,err := panFile[0].Open();
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening PAN file: %s", err.Error()))
			return
		}
		aadharHeader,err := aadharFile[0].Open();
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening Aadhar file: %s", err.Error()))
			return
		}



		panFileUrl,err := saveFile(panHeader,panFile[0]);
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving PAN file: %s", err.Error()))
			return
		}
		aadharFileUrl,err := saveFile(aadharHeader,aadharFile[0]);
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving Aadhar file: %s", err.Error()))
			return
		}

		
		defer panHeader.Close()
		defer aadharHeader.Close()

		seller.MobileNo = mobileno
		seller.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		seller.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		seller.Password = password
		seller.Company_Name = Company_Name
		seller.CompanyDetail.NameOfOwner = NameOfOwner
		seller.CompanyDetail.AadharNumber = AadharNumber
		seller.CompanyDetail.PAN = PAN
		seller.CompanyDetail.PermanentAddress = PermanentAddress
		seller.Email = email
		
		seller.CompanyDetail.AadharImage = aadharFileUrl
		seller.CompanyDetail.PANImage = panFileUrl
		
		seller.ID = primitive.NewObjectID()
		seller.Seller_ID = seller.ID.Hex()
		token, refreshtoken, _ := generate.TokenGenerator(seller.Email, seller.MobileNo, seller.Company_Name, seller.Seller_ID)
		seller.Token = token
		seller.Refresh_token = refreshtoken
		seller.User_type = "SELLER"
		validationErr := validate.Struct(seller)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": validationErr.Error()})
			return
		}
		_, inserterr := SellerCollection.InsertOne(ctx, seller)
		if inserterr != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating seller details: %s", err.Error()))
			return
		}

		c.String(http.StatusOK, "Seller details updated successfully!")

	}

}

// seller id is mandatory field to call this api
func ApproveSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		seller_id := c.PostForm("sellerid")
		if seller_id == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "seller id is not provided"})
		}
		sellerID, err := primitive.ObjectIDFromHex(seller_id)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "please provide valid seller id"})
			log.Fatal(err)
		}

		isApproved := c.PostForm("approved")
		var approved bool
		if isApproved == "approved" {
			approved = true
		}

		filter := primitive.M{utils.BsonID: sellerID}
		update := bson.M{
			"$set": bson.M{
				"approved": approved,
			}}
		_, err = SellerCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
		}
		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "updated successfully"})
	}
}








/* Login functions */



// pass form data
func SendLoginOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		contactNo := c.PostForm("mobileno")
		if contactNo == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "phone number can't be empty"})
			c.Abort()
			return
		}
		var founduser models.Seller
		err := SellerCollection.FindOne(ctx, bson.M{"mobileno": contactNo}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "No user exists with this phone number"})
			return
		}
		mobileNo := founduser.MobileNo
		otp, errG := generateOTP(mobileNo)
		if errG != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
			c.Abort()
			return
		}
		filter := primitive.M{utils.Mobileno: contactNo}
		update := primitive.M{
			"$set": primitive.M{
				"otp": otp,
			},
		}
		SellerCollection.UpdateOne(ctx, filter, update)
		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "OTP sent successfully"})

	}

}

// need to pass mobile no OTP and password
func LoginValidatePasswordOTP() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.Seller
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err})
			return
		}
		var founduser models.Seller

		err := SellerCollection.FindOne(ctx, bson.M{"mobileno": user.MobileNo}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error":"No account exist with this mobile."})
			return
		}

		passwordIsValid, msg := Verifypassword(user.Password, founduser.Password)
		if founduser.OTP != user.OTP || !passwordIsValid {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "OTP or password is incorrect"})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(founduser.Email, founduser.MobileNo, founduser.Company_Name, founduser.Seller_ID)
		generate.UpdateAllTokens(token, refreshToken, founduser.Seller_ID)
		c.JSON(http.StatusAccepted, token)

	}
}

func Verifypassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login or password is incorrect"
		valid = false
	}
	return valid, msg

}