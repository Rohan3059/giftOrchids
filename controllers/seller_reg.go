package controllers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
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
		count, err := SellerCollection.CountDocuments(ctx, filter)
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

func SellerEmailUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		//log current time
		currentTime := time.Now()
		fmt.Println("Current Time: ", currentTime)
		var seller models.Seller
		mobileno := c.PostForm("mobileno")
		email := c.PostForm("email")
		password := HashPassword(c.PostForm("password"))

		if mobileno == "" || email == "" || password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "All fields are required"})
			return
		}

		err := SellerCollection.FindOne(ctx, bson.M{"mobileno": mobileno}).Decode(&seller)

		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": " User already exists with this phone number"})
			return
		}

		err = SellerCollection.FindOne(ctx, bson.M{"email": email}).Decode(&seller)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User already exists with this email"})
			return
		}

		seller.ID = primitive.NewObjectID()
		seller.Seller_ID = seller.ID.Hex()
		seller.MobileNo = mobileno
		seller.Email = email
		seller.Password = HashPassword(password)
		seller.User_type = utils.Seller
		seller.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		seller.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		token, refreshtoken, _ := generate.TokenGenerator(seller.Email, seller.MobileNo, seller.Company_Name, seller.Seller_ID)
		seller.Token = token
		seller.Refresh_token = refreshtoken

		_, inserterr := SellerCollection.InsertOne(ctx, seller)
		finalTime := time.Now()
		fmt.Println("Current Time: ", currentTime)

		//total time taken
		fmt.Println("Total Time Taken: ", finalTime.Sub(currentTime))
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": inserterr.Error()})
			return
		}

		//log current time

		c.JSON(http.StatusOK, gin.H{"message": "Details updated sucessfully"})
	}
}

func SellerCommpanyDetailsUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var seller models.Seller
		mobileno := c.PostForm("mobileno")
		if mobileno == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Phone number can't be empty"})
			c.Abort()
			return
		}

		err := SellerCollection.FindOne(ctx, bson.M{"mobileno": mobileno}).Decode(&seller)
		if err != nil {
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
		GSTIN := c.PostForm("gstin")
		CIN := c.PostForm("cin")
		BusinessEntity := c.PostForm("businessentity")
		NoOfEmployee := c.PostForm("noofemployee")
		LLPIN := c.PostForm("llpin")

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		panFile := form.File["panFile"]
		gstinFile := form.File["gstinFile"]
		profile_picture := form.File["profile_picture"]

		if len(panFile) == 0 || len(gstinFile) == 0 || len(profile_picture) == 0 {
			c.String(http.StatusBadRequest, "Please upload all required documents")
			return
		}

		panHeader, err := panFile[0].Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening PAN file: %s", err.Error()))
			return
		}
		gstinHeader, err := gstinFile[0].Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening Aadhar file: %s", err.Error()))
			return
		}

		profile_Picture, err := profile_picture[0].Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening profile picture: %s", err.Error()))
			return
		}

		profilePicture_Url, err := saveFile(profile_Picture, profile_picture[0])
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving profile picture: %s", err.Error()))
			return
		}

		panFileUrl, err := saveFile(panHeader, panFile[0])
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving PAN file: %s", err.Error()))
			return
		}

		gstinFileUrl, err := saveFile(gstinHeader, gstinFile[0])
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving GSTIN file: %s", err.Error()))
			return
		}

		if LLPIN != "" {
			LLPINFile := form.File["llpinFile"]
			LLPINHeader, err := LLPINFile[0].Open()
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening LLPIN file: %s", err.Error()))
				return
			}
			LLPINFileUrl, err := saveFile(LLPINHeader, LLPINFile[0])
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving PAN file: %s", err.Error()))
				return
			}
			seller.CompanyDetail.LLPIN = LLPINFileUrl

			defer LLPINHeader.Close()

		}

		if CIN != "" {
			CINFile := form.File["cinFile"]
			CINHeader, err := CINFile[0].Open()
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening CIN file: %s", err.Error()))
				return
			}
			CINFileUrl, err := saveFile(CINHeader, CINFile[0])
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving CIN file: %s", err.Error()))
				return
			}
			seller.CompanyDetail.CIN = CINFileUrl
			defer CINHeader.Close()
		}

		defer panHeader.Close()
		defer gstinHeader.Close()

		defer profile_Picture.Close()

		seller.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		seller.Company_Name = Company_Name
		seller.CompanyDetail.PAN = PAN
		seller.CompanyDetail.PermanentAddress = PermanentAddress
		seller.CompanyDetail.ProfilePicture = profilePicture_Url
		seller.CompanyDetail.BusinessType = BusinessType
		seller.CompanyDetail.YearEstablished = YearEstablished
		seller.CompanyDetail.CompanyOrigin = CompanyOrigin
		seller.CompanyDetail.GSTIN = GSTIN
		seller.CompanyDetail.CIN = CIN
		if LLPIN != "" {
			seller.CompanyDetail.LLPIN = LLPIN
		}
		seller.CompanyDetail.BusinessEntity = BusinessEntity
		seller.CompanyDetail.NoOfEmployee = NoOfEmployee
		seller.CompanyDetail.PANImage = panFileUrl
		seller.CompanyDetail.GSTINDoc = gstinFileUrl

		filter := primitive.M{
			"mobileno": mobileno,
		}
		update := bson.M{"$set": bson.M{
			"company_name":  seller.Company_Name,
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

func SellerOwnerDetailsUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var seller models.Seller

		mobileno := c.PostForm("mobileno")
		if mobileno == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Mobile number is required"})
			return
		}

		OwnerName := c.PostForm("name")
		OwnerEmail := c.PostForm("email")
		OwnerMobileNo := c.PostForm("mobileno")
		OwnerGender := c.PostForm("gender")
		dob := c.PostForm("dateofbirth")
		aadharNumber := c.PostForm("aadharNumber")
		pan := c.PostForm("pan")
		havePassport, err := strconv.ParseBool(c.PostForm("havepassport"))

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

		aadharDocFile, err := aadharDoc[0].Open()
		if err != nil {
			log.Println("error while opening file")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process aadhar document"})
			return
		}
		defer aadharDocFile.Close()
		panDocFile, err := panDoc[0].Open()
		if err != nil {
			log.Println("error while opening file")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process PAN document"})
			return
		}
		defer panDocFile.Close()

		aadharDocUrl, saveError := saveFile(aadharDocFile, aadharDoc[0])
		if saveError != nil {

			c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving aadharDoc document"})
			return
		}

		panDocUrl, saveError := saveFile(panDocFile, panDoc[0])
		if saveError != nil {

			c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving panDoc document"})
			return
		}

		if passportDoc != nil {
			passportDocFile, err := passportDoc[0].Open()
			if err != nil {
				log.Println("error while opening file")

				c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to process passport document"})
				return
			}
			defer passportDocFile.Close()
			if passportDocFile != nil {
				passportDocUrl, saveError := saveFile(passportDocFile, passportDoc[0])
				if saveError != nil {

					c.JSON(http.StatusServiceUnavailable, gin.H{"Error": "Something went wrong while saving passportDoc document"})
					return
				}

				seller.OwnerDetail.PassportDocument = passportDocUrl
			}

		}

		seller.OwnerDetail.Name = OwnerName
		seller.OwnerDetail.Email = OwnerEmail
		seller.OwnerDetail.MobileNo = OwnerMobileNo
		seller.OwnerDetail.Gender = OwnerGender
		seller.OwnerDetail.DateOfBirth = dob
		seller.OwnerDetail.AadharNumber = aadharNumber
		seller.OwnerDetail.PAN = pan
		seller.OwnerDetail.HavePassport = havePassport
		seller.OwnerDetail.PassportNo = passportNo
		seller.OwnerDetail.AadharDocument = aadharDocUrl
		seller.OwnerDetail.PanDocument = panDocUrl

		filter := primitive.M{
			"mobileno": mobileno,
		}

		update := bson.M{"$set": bson.M{
			"ownerdetail": seller.OwnerDetail,
		},
		}

		_, updateError := SellerCollection.UpdateOne(ctx, filter, update)
		if updateError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to save owner details, try again "})
			return
		}

		c.String(http.StatusOK, "Owner details updated successfully!")

	}
}

func SellerLicenseUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var seller models.Seller

		mobileno := c.PostForm("mobileno")
		if mobileno == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Mobile number is required"})
			return
		}

		HaveBusinessLicensesStr := c.PostForm("havebusinesslicenses")
		HaveExportPermissionStr := c.PostForm("haveexportpermission")

		// Handle multipart form
		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while parsing multipart form:", err)
			c.String(http.StatusBadRequest, "error while parsing multipart form: %s", err.Error())
			return
		}

		business_LicenseFile := form.File["business_LicenseFile"]
		export_PermissionFile := form.File["export_PermissionFile"]

		HaveBusinessLicenses, err := strconv.ParseBool(HaveBusinessLicensesStr)
		if err != nil {
			log.Println("error parsing havebusinesslicenses:", err)
			HaveBusinessLicenses = false
		}

		HaveExportPermission, err := strconv.ParseBool(HaveExportPermissionStr)
		if err != nil {
			log.Println("error parsing haveexportpermission:", err)
			HaveExportPermission = false
		}

		var BusinessLicenseArray []models.BusinessLicense
		LicenseNameArray := c.PostFormArray("businessLicensename")
		LicenseValueArray := c.PostFormArray("businessLicensevalue")
		LicenseIssuedDateArray := c.PostFormArray("businessLicense_Issueddate")

		if HaveBusinessLicenses {
			if len(business_LicenseFile) == 0 {
				c.String(http.StatusBadRequest, "No business license files provided")
				return
			}
		}

		if len(LicenseNameArray) == 0 || len(LicenseValueArray) == 0 || len(LicenseIssuedDateArray) == 0 {
			c.String(http.StatusBadRequest, "No business license files provided")
			return
		}

		business_LicenseFileUrl := getFileUploadArray(business_LicenseFile)

		if len(business_LicenseFileUrl) == 0 {
			c.String(http.StatusBadRequest, "No business license files provided")
			return
		}

		for i := 0; i < len(LicenseNameArray); i++ {

			var licenseValue string
			if i < len(LicenseValueArray) {
				licenseValue = LicenseValueArray[i]
			} else {
				licenseValue = "NA"
			}

			var issuedDate string
			if i < len(LicenseIssuedDateArray) {
				issuedDate = LicenseIssuedDateArray[i]
			} else {
				issuedDate = "NA"
			}

			var licenseFileUrl string
			if i < len(business_LicenseFileUrl) {
				licenseFileUrl = business_LicenseFileUrl[i]
			} else {
				licenseFileUrl = ""
			}

			// Create BusinessLicense object
			BusinessLicenseArray = append(BusinessLicenseArray, models.BusinessLicense{
				LicenseName:  LicenseNameArray[i],
				LicenseValue: licenseValue,
				IssuedDate:   issuedDate,
				LicenseFile:  licenseFileUrl,
			})
		}

		var export_PermissionFileArray []models.BusinessLicense

		ExportLicenseArray := c.PostFormArray("exportPermissionname")
		ExportLicenseValueArray := c.PostFormArray("exportPermissionvalue")
		ExportLicenseIssuedDateArray := c.PostFormArray("exportPermission_Issueddate")
		export_PermissionFileUrl := getFileUploadArray(export_PermissionFile)

		if HaveExportPermission {
			if len(export_PermissionFile) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "No export permission files provided"})
				return
			}

			if len(ExportLicenseArray) == 0 || len(ExportLicenseValueArray) == 0 || len(ExportLicenseIssuedDateArray) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "No export permission files provided"})
				return
			}
		}

		for i := 0; i < len(ExportLicenseArray); i++ {
			// Check if ExportLicenseValueArray has enough elements
			var licenseValue string
			if i < len(ExportLicenseValueArray) {
				licenseValue = ExportLicenseValueArray[i]
			} else {
				licenseValue = "NA"
			}

			var issuedDate string
			if i < len(ExportLicenseIssuedDateArray) {
				issuedDate = ExportLicenseIssuedDateArray[i]
			} else {
				issuedDate = "NA"
			}

			// Check if export_PermissionFileUrl has enough elements
			var licenseFileUrl string
			if i < len(export_PermissionFileUrl) {
				licenseFileUrl = export_PermissionFileUrl[i]
			} else {
				licenseFileUrl = ""
			}

			// Create BusinessLicense object
			export_PermissionFileArray = append(export_PermissionFileArray, models.BusinessLicense{
				LicenseName:  ExportLicenseArray[i],
				LicenseValue: licenseValue,
				IssuedDate:   issuedDate,
				LicenseFile:  licenseFileUrl,
			})

		}

		seller.CompanyDetail.HaveBusinessLicenses = HaveBusinessLicenses
		seller.CompanyDetail.HaveExportPermission = HaveExportPermission
		seller.CompanyDetail.BusinessLicenses = BusinessLicenseArray
		seller.CompanyDetail.ExportPermission = export_PermissionFileArray

		filter := bson.M{"mobileno": mobileno}

		update := bson.M{"$set": bson.M{
			"companydetail.HaveBusinessLicenses": seller.CompanyDetail.HaveBusinessLicenses,
			"companydetail.HaveExportPermission": seller.CompanyDetail.HaveExportPermission,
			"companydetail.BusinessLicenses":     seller.CompanyDetail.BusinessLicenses,
			"companydetail.ExportPermission":     seller.CompanyDetail.ExportPermission,
		}}

		// Execute update operation
		_, updateErr := SellerCollection.UpdateOne(ctx, filter, update)
		if updateErr != nil {
			log.Println("error updating seller details:", updateErr)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating seller details: %s", updateErr.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Seller details updated successfully!"})
	}
}

func getFileUploadArray(fileHeaders []*multipart.FileHeader) []string {

	var fileUploadArray []string
	for _, file := range fileHeaders {
		f, err := file.Open()
		if err != nil {
			log.Println("error opening file:", err)
			return nil
		}
		defer f.Close()

		// Assuming saveFile is a function to save the file and get its URL
		fileUploadUrl, err := saveFile(f, file)
		if err != nil {
			log.Println("error saving file:", err)
			return nil
		}
		fileUploadArray = append(fileUploadArray, fileUploadUrl)
	}
	return fileUploadArray
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
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "No account exist with this mobile."})
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
	fmt.Println(err)
	valid := true
	msg := ""
	if err != nil {
		msg = "Login or password is incorrect"
		valid = false
	}
	return valid, msg

}
