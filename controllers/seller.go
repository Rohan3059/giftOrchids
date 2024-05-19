package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var SellerCollection *mongo.Collection = database.ProductData(database.Client, "seller")
var ProductReference *mongo.Collection = database.ProductData(database.Client, "ProductReference")

//var  *mongo.Collection = database.ProductData(database.Client, "seller")

// get all seller if no id is passesed all details if id id passed it will return sepcific seller
func GetSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Check if the user is a seller
		if checkSeller(ctx, c) {
			var sellerDetail models.Seller
			uid, _ := c.Get("uid")
			uids := fmt.Sprintf("%v", uid)
			sellerID, _ := primitive.ObjectIDFromHex(string(uids))
			filter := primitive.M{"_id": sellerID}
			if err := SellerCollection.FindOne(ctx, filter).Decode(&sellerDetail); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seller details"})
				return
			}
			c.JSON(http.StatusOK, sellerDetail)
			return
		}

		// Check if the user is an admin
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		// Parse seller ID from query
		sellerID := c.Query("sellerId")
		if sellerID != "" {
			// Fetch details of the specific seller
			var sellerDetail models.Seller
			sellerObjID, err := primitive.ObjectIDFromHex(sellerID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
				return
			}
			filter := primitive.M{"_id": sellerObjID}
			if err := SellerCollection.FindOne(ctx, filter).Decode(&sellerDetail); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found"})
				return
			}

			if sellerDetail.CompanyDetail.PANImage != "" {
				panPresignURL, err := getPresignURL(sellerDetail.CompanyDetail.PANImage)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.CompanyDetail.PANImage = panPresignURL
			}

			if sellerDetail.CompanyDetail.GSTINDoc != "" {
				gstPresignURL, err := getPresignURL(sellerDetail.CompanyDetail.GSTINDoc)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.CompanyDetail.GSTINDoc = gstPresignURL
			}

			if sellerDetail.CompanyDetail.ProfilePicture != "" {
				profilePresignURL, err := getPresignURL(sellerDetail.CompanyDetail.ProfilePicture)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.CompanyDetail.ProfilePicture = profilePresignURL
			}

			if sellerDetail.CompanyDetail.CINDoc != "" {
				ciPresignURL, err := getPresignURL(sellerDetail.CompanyDetail.CINDoc)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.CompanyDetail.CINDoc = ciPresignURL
			}

			if sellerDetail.CompanyDetail.LLPINDoc != "" {
				llpPresignURL, err := getPresignURL(sellerDetail.CompanyDetail.LLPINDoc)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.CompanyDetail.LLPINDoc = llpPresignURL
			}

			if sellerDetail.OwnerDetail.AadharDocument != "" {
				aadharPresignURL, err := getPresignURL(sellerDetail.OwnerDetail.AadharDocument)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.OwnerDetail.AadharDocument = aadharPresignURL
			}

			if sellerDetail.OwnerDetail.PanDocument != "" {
				panPresignURL, err := getPresignURL(sellerDetail.OwnerDetail.PanDocument)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.OwnerDetail.PanDocument = panPresignURL
			}

			if sellerDetail.OwnerDetail.PassportDocument != "" {
				passportPresignURL, err := getPresignURL(sellerDetail.OwnerDetail.PassportDocument)
				if err != nil {
					log.Println(err)
				}
				sellerDetail.OwnerDetail.PassportDocument = passportPresignURL
			}

			c.JSON(http.StatusOK, sellerDetail)
			return
		}

		// Fetch details of all sellers
		var sellerDetails []models.Seller
		cur, err := SellerCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to find sellers"})
			return
		}
		if err := cur.All(ctx, &sellerDetails); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seller details"})
			return
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

func AddProductReferenceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			SellerID    string `json:"seller_id" binding:"required"`
			ProductID   string `json:"product_id" binding:"required"`
			Price       string `json:"price" binding:"required"`
			MinQuantity int    `json:"min_quantity" binding:"required"`
			MaxQuantity int    `json:"max_quantity" binding:"required"`
		}
		ctx := context.Background()

		if err := c.ShouldBindJSON(&input); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sellerID, err := primitive.ObjectIDFromHex(input.SellerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seller ID"})
			return
		}

		productID, err := primitive.ObjectIDFromHex(input.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}

		productReference := models.ProductReference{
			ID:          primitive.NewObjectID(),
			ProductID:   productID,
			SellerID:    sellerID,
			Price:       input.Price,
			MinQuantity: input.MinQuantity,
			MaxQuantity: input.MaxQuantity,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Approved:    false, // You may set the default value as needed
			Archived:    false, // You may set the default value as needed
		}

		// Insert product reference into the ProductReferenceCollection
		_, err = ProductReference.InsertOne(ctx, productReference)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update seller model with the product reference ID
		update := bson.M{"$push": bson.M{"product_references": productReference.ID}}
		_, err = SellerCollection.UpdateOne(ctx, bson.M{"_id": sellerID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update product model with the product reference ID
		updateProduct := bson.M{"$push": bson.M{"product_references": productReference.ID}}
		_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": productID}, updateProduct)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Product reference added successfully"})
	}
}

func SellerUpdateProfilePictureHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sellerID, exists := c.Get("uid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Seller ID not found in context"})
			return
		}

		// Convert sellerID to ObjectID
		sellerObjID, err := primitive.ObjectIDFromHex(sellerID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Seller ID"})
			return
		}

		// Query the database to get seller information
		var seller models.Seller // Assuming Seller struct is defined in models package
		err = SellerCollection.FindOne(context.Background(), bson.M{"_id": sellerObjID}).Decode(&seller)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Seller not found"})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		profile_picture := form.File["profile_picture"]

		// check if files are present
		if len(profile_picture) == 0 {
			c.String(http.StatusBadRequest, "Please upload all required documents")
			return
		}

		profilePicture, err := profile_picture[0].Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening profile picture: %s", err.Error()))
			return
		}

		profilePictureUrl, err := saveFile(profilePicture, profile_picture[0])
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving profile picture: %s", err.Error()))
			return
		}
		fmt.Println(profilePictureUrl)

		filter := bson.M{"_id": sellerObjID}
		update := bson.M{"$set": bson.M{"companydetail.profilepicture": profilePictureUrl}}
		_, err = SellerCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}
		c.Header("content-type", "application/json")
		c.JSON(http.StatusOK, gin.H{"success": "Profile Picture updated successfully"})

	}
}

// find all products for specifc seller stored in sellerRegistered array
func GetAllProductsForASellerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		uid, exist := c.Get("uid")
		if !exist {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{"Error": "You're not authroized to perform this action"})
		}
		var sellerId = uid.(string)

		var products []models.Product

		cursor, err := ProductCollection.Find(ctx, bson.M{"sellerregistered": sellerId})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = cursor.All(ctx, &products)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//itertate through all products and generate presign url for image

		for i := 0; i < len(products); i++ {

			if products[i].Image != nil {

				for j := 0; j < len(products[i].Image); j++ {

					imageUrl, err := getPresignURL(products[i].Image[j])

					if err !=

						nil {

						c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})

						return

					}

					products[i].Image[j] = imageUrl

				}

			}
		}

		c.JSON(http.StatusOK, products)

	}

}

func UpdateSellerBusinessDetails() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, exist := c.Get("uid")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You're not authorized to perform this action"})
			return
		}

		sellerId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Seller ID"})
			return
		}

		var seller models.Seller

		findErr := SellerCollection.FindOne(ctx, bson.M{"_id": sellerId}).Decode(&seller)
		if findErr != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to find seller with this phone number"})
			c.Abort()
			return
		}

		Company_Name := c.PostForm("Company_name")
		PAN := c.PostForm("pan")
		PermanentAddress := c.PostForm("permanenetaddress")

		BusinessType := c.PostForm("businesstype")
		YearEstablished := c.PostForm("yearestablished")
		CompanyOrigin := c.PostForm("companyorigin")
		GSTINORCIN := c.PostForm("gstinorcin")
		BusinessEntity := c.PostForm("businessentity")
		NoOfEmployee := c.PostForm("noofemployee")

		seller.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		seller.Company_Name = Company_Name
		seller.CompanyDetail.PAN = PAN
		seller.CompanyDetail.PermanentAddress = PermanentAddress
		seller.CompanyDetail.BusinessType = BusinessType
		seller.CompanyDetail.YearEstablished = YearEstablished
		seller.CompanyDetail.CompanyOrigin = CompanyOrigin
		seller.CompanyDetail.CIN = GSTINORCIN
		seller.CompanyDetail.BusinessEntity = BusinessEntity
		seller.CompanyDetail.NoOfEmployee = NoOfEmployee

		validationErr := validate.Struct(seller)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": validationErr.Error()})
			return
		}

		filter := primitive.M{
			"_id": sellerId,
		}
		update := bson.M{"$set": bson.M{
			"Company_name":  seller.Company_Name,
			"companydetail": seller.CompanyDetail,
		},
		}

		_, updateError := SellerCollection.UpdateOne(ctx, filter, update)

		if updateError != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": updateError.Error()})
			return
		}

		c.String(http.StatusOK, "Seller details updated successfully!")

	}
}

// update owner details
func UpdateOwnerDetails() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, exist := c.Get("uid")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You're not authorized to perform this action"})
			return
		}

		sellerId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Seller"})
			return
		}

		var seller models.Seller

		findErr := SellerCollection.FindOne(ctx, bson.M{"_id": sellerId}).Decode(&seller)
		if findErr != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to find seller with this phone number"})
			c.Abort()
			return
		}

		update := bson.M{}

		OwnerName := c.PostForm("name")
		OwnerEmail := c.PostForm("email")
		OwnerMobileNo := c.PostForm("mobileno")
		OwnerGender := c.PostForm("gender")
		dob := c.PostForm("dateofbirth")
		aadharNumber := c.PostForm("aadharNumber")
		pan := c.PostForm("pan")
		passportNo := c.PostForm("passportNo")

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}

		aadharDoc := form.File["aadharDoc"]
		panDoc := form.File["panDoc"]
		passportDoc := form.File["passportDoc"]

		var aadharDocUrl string
		var panDocUrl string
		var passportDocUrl string

		if len(aadharDoc) > 0 {
			aadharDocFile, err := aadharDoc[0].Open()
			if err != nil {
				log.Println("error while opening file")
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process aadhar document"})
				return
			}
			defer aadharDocFile.Close()
			url, saveError := saveFile(aadharDocFile, aadharDoc[0])
			if saveError != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving aadharDoc document"})
				return
			}
			aadharDocUrl = url
		}

		if len(panDoc) > 0 {
			panDocFile, err := panDoc[0].Open()
			if err != nil {
				log.Println("error while opening file")
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process PAN document"})
				return
			}
			defer panDocFile.Close()
			url, saveError := saveFile(panDocFile, panDoc[0])
			if saveError != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving panDoc document"})
				return
			}
			panDocUrl = url
		}

		if len(passportDoc) > 0 {
			passportDocFile, err := passportDoc[0].Open()
			if err != nil {
				log.Println("error while opening file")
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process passport document"})
				return
			}
			defer passportDocFile.Close()
			url, saveError := saveFile(passportDocFile, passportDoc[0])
			if saveError != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving passportDoc document"})
				return
			}
			passportDocUrl = url
		}

		if OwnerName != "" {
			update["ownerdetail.name"] = OwnerName
		}

		if OwnerEmail != "" {
			update["ownerdetail.email"] = OwnerEmail
		}

		if OwnerMobileNo != "" {
			update["ownerdetail.mobileno"] = OwnerMobileNo
		}

		if OwnerGender != "" {
			update["ownerdetail.gender"] = OwnerGender
		}

		if dob != "" {
			update["ownerdetail.dateofbirth"] = dob
		}

		if aadharNumber != "" {
			update["ownerdetail.aadharNumber"] = aadharNumber
		}

		if pan != "" {
			update["ownerdetail.pan"] = pan
		}

		if passportNo != "" {
			update["ownerdetail.passportNo"] = passportNo
			update["ownerdetail.havepassport"] = true
		}

		if aadharDocUrl != "" {
			update["ownerdetail.aadharDoc"] = aadharDocUrl
		}

		if panDocUrl != "" {
			update["ownerdetail.panDoc"] = panDocUrl
		}

		if passportDocUrl != "" {
			update["ownerdetail.passportDoc"] = passportDocUrl
		}

		update["approved"] = false

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		update["updated_at"] = updated_at

		if len(update) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "No valid fields to update"})
			return
		}

		filter := bson.M{"_id": sellerId}
		fmt.Println(update)

		updateError := SellerCollection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}).Decode(&seller)
		if updateError != nil {
			if errors.Is(updateError, mongo.ErrNoDocuments) {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "No seller found"})
				return
			}

			log.Println(updateError)
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to save owner details, try again"})
			return
		}

		c.String(http.StatusOK, "Owner details updated successfully!")
	}
}

func SellerPasswordConfirmation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, exist := c.Get("uid")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
			return
		}

		sellerId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
			return
		}

		var seller models.Seller

		password := c.PostForm("password")

		filter := bson.M{"_id": sellerId}

		//match id and password
		err = SellerCollection.FindOne(ctx, filter).Decode(&seller)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
			return
		}

		validPassword, msg := Verifypassword(password, seller.Password)

		if !validPassword {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Incorrect Password"})
			fmt.Print(msg)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": true})

	}
}

func UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, exist := c.Get("uid")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
			return
		}

		sellerId, err := primitive.ObjectIDFromHex(id.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
			return
		}

		var seller models.Seller

		password := c.PostForm("password")
		hashPassword := HashPassword(password)
		newPassword := c.PostForm("new_password")
		newPasswordHash := HashPassword(newPassword)

		filter := bson.M{"_id": sellerId, "password": hashPassword}
		update := bson.M{"$set": bson.M{"password": newPasswordHash}}

		//match id and password
		err = SellerCollection.FindOneAndUpdate(ctx, filter, update).Decode(&seller)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Incorrect Password"})
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})

	}
}

func SellerOtpVerfication() gin.HandlerFunc {
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
		var seller models.Seller
		err := SellerCollection.FindOne(ctx, filter).Decode(&seller)

		if err != nil && err != mongo.ErrNoDocuments {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}
		if err == mongo.ErrNoDocuments {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Mobile number doesn't exist"})
			c.Abort()
			return
		}

		if seller.OTP == otp {
			SellerCollection.FindOneAndUpdate(ctx, filter, primitive.M{"otp": ""})
			c.Header("content-type", "application/json")
			c.JSON(http.StatusOK, gin.H{"success": "verified"})
		} else {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid OTP"})
		}
	}
}

func DownloadSellerDocs() gin.HandlerFunc {

	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		docType := c.Query("doc")
		id := c.Query("sellerId")

		if checkSeller(ctx, c) {

			uid, exist := c.Get("uid")

			if !exist {

				c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
				return
			}
			sellerId, err := primitive.ObjectIDFromHex(uid.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
				return
			}

			filter := bson.M{"_id": sellerId}

			var foundSeller models.Seller

			//find seller
			err = SellerCollection.FindOne(ctx, filter).Decode(&foundSeller)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})

			}

			byteRes, err := SellerDocDownload(c, sellerId, docType)

			if err != nil {

				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})

			}

			c.Data(http.StatusOK, "application/pdf", byteRes)

		} else {
			sellerId, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})
				return
			}
			byteRes, err := SellerDocDownload(c, sellerId, docType)

			if err != nil {

				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return

			}

			c.Data(http.StatusOK, "application/pdf", byteRes)

		}

	}
}

func SellerDocDownload(c *gin.Context, sellerId primitive.ObjectID, docType string) ([]byte, error) {

	var seller models.Seller

	filter := bson.M{"_id": sellerId}

	err := SellerCollection.FindOne(context.TODO(), filter).Decode(&seller)
	if err != nil {

		return nil, err
	}

	if docType == "aadhar" {

		aadharFile, err := DownloadPDFFromS3(seller.OwnerDetail.AadharDocument)

		return aadharFile, err

	}

	if docType == "owner_pan" {

		ownerPanFile, err := DownloadPDFFromS3(seller.OwnerDetail.PanDocument)

		return ownerPanFile, err
	}

	if docType == "company_pan" {

		companyPanFile, err := DownloadPDFFromS3(seller.CompanyDetail.PANImage)

		return companyPanFile, err
	}

	if docType == "gstin" {

		companyGstFile, err := DownloadPDFFromS3(seller.CompanyDetail.GSTINDoc)

		return companyGstFile, err
	}

	if docType == "cin" {

		companyCinFile, err := DownloadPDFFromS3(seller.CompanyDetail.CINDoc)

		return companyCinFile, err
	}

	if docType == "llpin" {

		companyLlpinFile, err := DownloadPDFFromS3(seller.CompanyDetail.LLPINDoc)

		return companyLlpinFile, err
	}

	return nil, errors.New("unable to download file")

}

func DownloadAllFiles() gin.HandlerFunc {

	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := c.Query("id")

		sellerId, err := primitive.ObjectIDFromHex(id)

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})

		}

		//find seller and all docs

		var seller models.Seller

		filter := bson.M{"_id": sellerId}

		err = SellerCollection.FindOne(ctx, filter).Decode(&seller)

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"Error": "You're not authorized to perform this action"})

		}

	}

}

func LoadSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		sellerID, exists := c.Get("uid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Seller ID not found in context"})
			return
		}

		sellerObjID, err := primitive.ObjectIDFromHex(sellerID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Seller ID"})
			return
		}

		// Query the database to get seller information
		var seller models.Seller // Assuming Seller struct is defined in models package
		err = SellerCollection.FindOne(context.Background(), bson.M{"_id": sellerObjID}).Decode(&seller)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found"})
			return
		}

		if seller.CompanyDetail.ProfilePicture != "" {
			profilePictureUrl, err := getPresignURL(seller.CompanyDetail.ProfilePicture)
			if err != nil {
				//
			}

			seller.CompanyDetail.ProfilePicture = profilePictureUrl
		}

		seller.CompanyDetail = getPresignUrlOfSellerBusinessDoc(seller.CompanyDetail)
		seller.OwnerDetail = getPresignUrlOfOwnerDocs(seller.OwnerDetail)

		if !seller.Approved {
			c.JSON(http.StatusOK, gin.H{"message": "Seller is not approved", "isApproved": false, "seller": seller})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Seller is approved", "isApproved": true, "seller": seller})
	}
}

func getPresignUrlOfSellerBusinessDoc(companyDetail models.CompanyDetail) models.CompanyDetail {

	if companyDetail.GSTINDoc != "" {

		url, err := getPresignURL(companyDetail.GSTINDoc)
		if err != nil {
			url = ""
		}

		companyDetail.GSTINDoc = url
	}

	if companyDetail.PANImage != "" {

		url, err := getPresignURL(companyDetail.PANImage)
		if err != nil {
			url = ""
		}

		companyDetail.PANImage = url
	}

	if companyDetail.CINDoc != "" {

		url, err := getPresignURL(companyDetail.CINDoc)
		if err != nil {
			url = ""
		}

		companyDetail.CINDoc = url
	}

	if companyDetail.LLPINDoc != "" {

		url, err := getPresignURL(companyDetail.LLPINDoc)
		if err != nil {
			url = ""
		}

		companyDetail.LLPINDoc = url
	}

	return companyDetail

}

func getPresignUrlOfOwnerDocs(ownerDetails models.OwnerDetails) models.OwnerDetails {

	if ownerDetails.PassportDocument != "" {

		url, err := getPresignURL(ownerDetails.PassportDocument)
		if err != nil {
			url = ""
		}

		ownerDetails.PassportDocument = url

	}

	if ownerDetails.AadharDocument != "" {

		url, err := getPresignURL(ownerDetails.AadharDocument)
		if err != nil {
			url = ""
		}

		ownerDetails.AadharDocument = url
	}

	if ownerDetails.PanDocument != "" {

		url, err := getPresignURL(ownerDetails.PanDocument)
		if err != nil {
			url = ""
		}
		ownerDetails.PanDocument = url
	}

	return ownerDetails

}

// delete image from a product based on index from query
func DeleteImageFromProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := c.Param("id")

		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "productID is required"})
			return
		}

		productID, _ := primitive.ObjectIDFromHex(id)
		i := c.Query("index")

		// Parse the query parameter to an integer
		index, err := strconv.Atoi(i)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid index"})
			return
		}

		var product models.Product

		finderr := ProductCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)

		if checkSeller(ctx, c) {
			uid, exist := c.Get("uid")
			if !exist {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "You are not authorized to perform this action "})
				return
			}

			//check product.sellerregistered array has objectid of seller
			var isAuthorized bool

			for _, v := range product.SellerRegistered {
				isAuthorized = false
				if v == uid.(string) {
					isAuthorized = true
				}
			}

			if !isAuthorized {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "You are not authorized to perform this action "})
				return
			}

		}

		if finderr != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "product not found"})
			return
		}

		product.Image = append(product.Image[:index], product.Image[index+1:]...)

		_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": productID}, bson.M{"$set": bson.M{"image": product.Image}})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to delete image"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "image deleted successfully"})

	}

}

func SellerUpdateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !checkSeller(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		if product.Product_ID.Hex() == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			c.Abort()
			return
		}
		if product.Product_ID.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Product ID"})
			c.Abort()
			return
		}
		filter := primitive.M{"_id": product.Product_ID}

		update := bson.M{}

		if product.Product_Name != "" {
			update["name"] = product.Product_Name
		}
		if product.Category != "" {
			update["category"] = product.Category
		}

		if product.Discription != "" {
			update["discription"] = product.Discription
		}

		if product.Price != "" {
			update["price"] = product.Price
		}

		if product.SKU != "" {
			update["sku"] = product.SKU
		}

		if product.Attributes != nil {
			update["attributes"] = product.Attributes
		}

		updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		update["updated_at"] = updated_at

		var newProduct models.Product
		err := ProductCollection.FindOneAndUpdate(ctx, filter, update).Decode(&newProduct)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
				return
			}
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, product)
	}

}

// get support ticket for specific seller
func GetSellerSupportTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if !checkSeller(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		uid, exist := c.Get("uid")

		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "You are not authorized to perform this action "})
			return
		}
		//convert to objectid get

		sellerID, _ := primitive.ObjectIDFromHex(uid.(string))

		filter := bson.M{"_id": sellerID}

		var seller models.Seller
		err := SellerCollection.FindOne(ctx, filter).Decode(&seller)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": "something went wrong"})
			return
		}

		mobileno := seller.MobileNo
		email := seller.Email

		support_filter := bson.M{
			"mobileno": mobileno,
			"email":    email,
		}

		var support []models.CustomerSupportTicket

		cursor, err := SupportTickerCollection.Find(ctx, support_filter)

		if err != nil {
			fmt.Println("Ticket error")
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": "something went wrong"})
			return
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var ticket models.CustomerSupportTicket
			err := cursor.Decode(&ticket)
			if err != nil {
				fmt.Println("Ticket decode error")
				fmt.Println(err)
				c.JSON(http.StatusBadRequest, gin.H{"Error": "something went wrong"})
				return
			}

			support = append(support, ticket)

		}

		c.JSON(http.StatusOK, support)
	}

}
