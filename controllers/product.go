package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var EnquireCollection *mongo.Collection = database.ProductData(database.Client, "enquire")
var RequirementMessageCollection *mongo.Collection = database.ProductData(database.Client, "RequirementMessage")
var ReviewsCollection *mongo.Collection = database.ProductData(database.Client, "Reviews")
var FeedsCollection *mongo.Collection = database.ProductData(database.Client, "New_Feeds")

var uploader *s3manager.Uploader

func init() {

	/*if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}*/

	// Read AWS credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	//load these from aws secret manager

	// Create a new AWS session with the provided credentials and region
	awsSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				accessKey,
				secretKey,
				"",
			),
		},
	})

	if err != nil {
		panic(err)
	}

	// Create an uploader instance using the AWS session
	uploader = s3manager.NewUploader(awsSession)
}

func saveFile(fileReader io.Reader, fileHeader *multipart.FileHeader) (string, error) {
	// Upload the file to S3 using the fileReader

	bucketName := os.Getenv("AWS_BUCKET_NAME")

	mtype, error := mimetype.DetectReader(fileReader)

	if error != nil {
		return "", error
	}

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fileHeader.Filename),
		Body:        fileReader,
		ContentType: aws.String(mtype.String()),
	})
	if err != nil {
		return "", err
	}

	// Get the URL of the uploaded file
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, fileHeader.Filename)

	return url, nil
}

func extractKeyFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func DownloadPDFFromS3(s3Url string) ([]byte, error) {
	// Create a new AWS session
	sess := session.Must(session.NewSession())
	bucketName := os.Getenv("AWS_BUCKET_NAME")

	// Create a downloader with the S3 client
	downloader := s3manager.NewDownloader(sess)

	keyName := extractKeyFromURL(s3Url)

	if keyName == "" {
		return nil, errors.New("Invalid S3 URL" + keyName)
	}

	// Input parameters for the S3 object you want to download
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}

	// Create a buffer to hold the downloaded file contents
	buffer := aws.NewWriteAtBuffer([]byte{})

	// Download the file from S3 and write it to the buffer
	_, err := downloader.Download(buffer, input)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func getPresignURL(s3Url string) (string, error) {
	// Create an S3 service client using the provided session

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	// Create a new AWS session with the provided credentials and region
	awsSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				accessKey,
				secretKey,
				"",
			),
		},
	})

	if err != nil {
		panic(err)
	}

	if awsSession == nil {
		fmt.Println("AWS session is nil")
		return "", nil
	}

	s3client := s3.New(awsSession)

	// Retrieve bucket name from environment variable
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	if bucketName == "" {
		fmt.Println("Bucket name is empty")
		return "", nil
	}

	keyName := extractKeyFromURL(s3Url)
	if keyName == "" {
		return "", nil // Return nil error as keyName is empty
	}

	var contentType string

	if strings.Contains(s3Url, "pdf") {
		contentType = "application/pdf"
	} else if strings.Contains(s3Url, "docx") {
		contentType = "application/docx"
	} else if strings.Contains(s3Url, "mp4") {
		contentType = "video/mp4"
	} else if strings.Contains(s3Url, "jpg") {
		contentType = "image/jpg"
	} else if strings.Contains(s3Url, "png") {
		contentType = "image/png"
	} else if strings.Contains(s3Url, "webp") {
		contentType = "image/webp"
	} else if strings.Contains(s3Url, "avif") {
		contentType = "image/avif"
	} else if strings.Contains(s3Url, "svg") {
		contentType = "image/svg"
	} else if strings.Contains(s3Url, "jpeg") {
		contentType = "image/jpeg"
	}

	req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                  aws.String(bucketName),
		Key:                     aws.String(keyName),
		ResponseContentType:     aws.String(contentType),
		ResponseContentEncoding: aws.String("base64"),
	})

	q := req.HTTPRequest.URL.Query()
	q.Add("x-amz-acl", "public-read")
	q.Add("Content-Type", contentType)
	req.HTTPRequest.URL.RawQuery = q.Encode()

	presignedURL, err := req.Presign(time.Hour * 24) // URL expires in 24 hours
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var sellerId string
		if checkSeller(ctx, c) {
			sellerID, exists := c.Get("uid")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Seller ID not found in context"})
				return
			}

			sellerId = sellerID.(string)

		}

		var product models.Product

		defer cancel()

		product.Product_ID = primitive.NewObjectID()
		product.Product_Name = c.PostForm("product_name")

		attributesList := c.PostForm("attributes")
		priceRanges := c.PostForm("priceRange")

		if priceRanges != "" {
			var productPriceRanges []models.ProductPriceRange
			if err := json.Unmarshal([]byte(priceRanges), &productPriceRanges); err != nil {
				fmt.Println(err)
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Error while parsing price range"})
				return
			}

			product.PriceRange = productPriceRanges
		}

		var attributes []models.AttributeValue
		if err := json.Unmarshal([]byte(attributesList), &attributes); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		product.Attributes = attributes

		price := strings.TrimSpace(c.PostForm("price"))

		product.Price = price

		product.Discription = c.PostForm("discription")
		product.Category = c.PostForm("category")
		//add time

		product.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		product.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]

		var errors []string
		var uploadedURLs []string
		var resFileName []string
		fmt.Println(files)
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				log.Fatal(err)
				log.Println("error while opening file")
				c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
				return
			}
			uploadedURL, err := saveFile(f, file)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error saving file %s: %s", file.Filename, err.Error()))
			} else {
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
			if len(errors) > 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errors})
			} else {
				c.JSON(http.StatusOK, gin.H{"url": uploadedURLs})
			}
		}
		fmt.Println(resFileName)
		product.Image = uploadedURLs

		var sellers []string
		sellers = append(sellers, sellerId)
		product.SellerRegistered = sellers
		if sellerId != "" {
			product.AddedBy = "seller"
		} else {
			product.AddedBy = "Admin"
		}

		_, anyerr := ProductCollection.InsertOne(ctx, product)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added our Product Admin!!")
	}
}

func AddProductByAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var product models.Product

		defer cancel()

		sellerId := c.PostForm("sellerId")
		if sellerId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Seller id is required"})
			return
		}
		product.Product_ID = primitive.NewObjectID()
		product.Product_Name = c.PostForm("product_name")

		attributesList := c.PostForm("attributes")
		priceRanges := c.PostForm("priceRange")
		fmt.Println(priceRanges)

		if priceRanges != "" {
			var productPriceRanges []models.ProductPriceRange
			if err := json.Unmarshal([]byte(priceRanges), &productPriceRanges); err != nil {
				fmt.Println(err)
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Error while parsing price range"})
				return
			}

			product.PriceRange = productPriceRanges
		}

		var attributes []models.AttributeValue
		if err := json.Unmarshal([]byte(attributesList), &attributes); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		product.Attributes = attributes

		price := strings.TrimSpace(c.PostForm("price"))

		product.Price = price

		product.Discription = c.PostForm("discription")
		product.Category = c.PostForm("category")
		//add time

		product.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		product.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]

		var errors []string
		var uploadedURLs []string
		var resFileName []string
		fmt.Println(files)
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				log.Fatal(err)
				log.Println("error while opening file")
				c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
				return
			}
			uploadedURL, err := saveFile(f, file)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error saving file %s: %s", file.Filename, err.Error()))
			} else {
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
			if len(errors) > 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errors})
			} else {
				c.JSON(http.StatusOK, gin.H{"url": uploadedURLs})
			}
		}
		fmt.Println(resFileName)
		product.Image = uploadedURLs

		var sellers []string
		sellers = append(sellers, sellerId)
		product.SellerRegistered = sellers
		product.AddedBy = "Admin"

		_, anyerr := ProductCollection.InsertOne(ctx, product)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added our Product Admin!!")
	}
}

// this will give detail of the particular product, product id is mendatory filed
func GetProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		productID := c.Query("productId")
		if productID == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			c.Abort()
			return
		}
		prodID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			c.Abort()
			return
		}

		type ProductWithAttributes struct {
			models.Product
			AttributesInfo []models.AttributeType `bson:"attributes_info" json:"attributes_info"`
		}

		// Perform aggregation
		pipeline := []bson.M{
			{
				"$match": bson.M{"_id": prodID},
			},
			{
				"$lookup": bson.M{
					"from":         "AttributeType",
					"localField":   "attributes.attribute_type",
					"foreignField": "_id",
					"as":           "attributes_info",
				},
			},
		}

		// Create a variable to store the result
		var result ProductWithAttributes

		// Perform aggregation
		cursor, err := ProductCollection.Aggregate(ctx, pipeline)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		// Iterate over the cursor and decode the result
		// Iterate over the cursor and decode the result
		for cursor.Next(ctx) {
			if err := cursor.Decode(&result); err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
			// Decode Product details separately
			if err := cursor.Decode(&result.Product); err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
		}

		// Check if no result found
		if result.Product.Product_ID.IsZero() {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
			return
		}

		fmt.Println(result.Product.Image)
		fmt.Println(len(result.Product.Image))

		if result.Product.Image != nil {
			for i, url := range result.Product.Image {
				imageUrl, err := getPresignURL(url)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
					return
				}
				result.Product.Image[i] = imageUrl
			}

		}

		c.JSON(http.StatusOK, result)
	}

}

// this will update product need to pass whole struct json data
func UpdateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "You are not authorized to perform this action."})
			return
		}

		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Product ID is required"})
			return
		}

		productID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.M{"_id": productID}

		var product models.Product

		err := ProductCollection.FindOne(ctx, filter).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
			return
		}

		product_name := c.PostForm("product_name")

		attributesList := c.PostForm("attributes")
		priceRanges := c.PostForm("priceRange")

		var productPriceRanges []models.ProductPriceRange
		if priceRanges != "" {
			if err := json.Unmarshal([]byte(priceRanges), &productPriceRanges); err != nil {
				fmt.Println(err)

			}
		}

		var attributes []models.AttributeValue
		if attributesList != "" {
			if err := json.Unmarshal([]byte(attributesList), &attributes); err != nil {
				fmt.Println(err)
			}
		}

		price := strings.TrimSpace(c.PostForm("price"))
		description := c.PostForm("discription")
		category := c.PostForm("category")
		sku := c.PostForm("sku")
		updated_at := time.Now()

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]

		var errors []string
		var uploadedURLs []string
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				log.Fatal(err)
				c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
				return
			}
			uploadedURL, err := saveFile(f, file)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error saving file %s: %s", file.Filename, err.Error()))
			} else {
				uploadedURLs = append(uploadedURLs, uploadedURL)
			}
		}
		if len(errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors})
			return
		}

		update := bson.M{}
		if product_name != "" {
			update["product_name"] = product_name
		}

		if sku != "" {
			update["sku"] = sku
		}

		if price != "" {
			update["price"] = price
		}
		if description != "" {
			update["discription"] = description
		}
		if category != "" {
			update["category"] = category
		}
		update["updated_at"] = updated_at

		pushUpdates := bson.M{}
		if len(uploadedURLs) > 0 {
			pushUpdates["image"] = bson.M{
				"$each": uploadedURLs,
			}
		}
		if len(productPriceRanges) > 0 {
			pushUpdates["pricerange"] = bson.M{
				"$each": productPriceRanges,
			}
		}
		if len(attributes) > 0 {
			pushUpdates["attributes"] = bson.M{
				"$each": attributes,
			}
		}

		updateOperation := bson.M{"$set": update}
		if len(pushUpdates) > 0 {
			updateOperation["$push"] = pushUpdates
		}

		result, err := ProductCollection.UpdateOne(ctx, filter, updateOperation)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while updating product"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Message": "Product updated successfully"})
	}

}

// accept diff filter to get product
func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		queryParam := c.Query("name")
		filter := []primitive.M{}

		if queryParam != "" {
			filter = append(filter, primitive.M{"name": primitive.M{"$regex": queryParam, "$options": "i"}})
		}

		productCategory := c.Query("category")
		productCategory = strings.TrimSpace(productCategory)
		if productCategory != "" {
			filter = append(filter, primitive.M{utils.Categories: productCategory})
		}

		brand := c.Query("brand")
		if brand != "" {
			filter = append(filter, primitive.M{utils.Brand: brand})
		}

		productName := c.Query("productname")
		if productName != "" {
			filter = append(filter, primitive.M{utils.ProductName: productName})
		}

		filter = append(filter, primitive.M{"approved": true})

		finalFilter := primitive.M{}
		if len(filter) > 0 {
			finalFilter = primitive.M{"$and": filter}
		}

		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil || limit <= 0 {
			limit = 20 // Default limit
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page <= 0 {
			page = 1 // Default page
		}

		skip := (page - 1) * limit

		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		findOptions := options.Find()
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(skip))
		findOptions.SetSort(bson.D{{Key: "updated_at", Value: -1}}) // Sort by updated_at in descending order

		cursor, err := ProductCollection.Find(ctx, finalFilter, findOptions)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Error fetching products: " + err.Error()})
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &searchProducts); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error decoding products: " + err.Error()})
			return
		}

		for i := range searchProducts {
			for j := range searchProducts[i].Image {
				url, err := getPresignURL(searchProducts[i].Image[j])
				if err != nil {
					log.Println("Error generating pre-signed URL for image:", err)
					continue
				}
				searchProducts[i].Image[j] = url
			}
		}

		//find if it has more products to be fetched

		count, err := ProductCollection.CountDocuments(ctx, finalFilter)
		if err != nil {

			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Error fetching products"})
			return

		}

		c.IndentedJSON(http.StatusOK, gin.H{
			"products": searchProducts,
			"page":     page,
			"limit":    limit,
			"nextPage": page + 1,
			"hasMore":  count > int64(page*limit),
		})

	}
}

func GetAllProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.M{})
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Something went wrong while fetching the data")
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &searchProducts); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "Invalid")
			return
		}

		// Iterate over each product and get pre-signed URLs for each image
		for i := range searchProducts {
			for j := range searchProducts[i].Image {
				// Get pre-signed URL for the image
				url, err := getPresignURL(searchProducts[i].Image[j])
				if err != nil {
					log.Println("Error generating pre-signed URL for image:", err)
					continue
				}
				// Update the image URL in the product
				searchProducts[i].Image[j] = url
			}
		}

		c.IndentedJSON(http.StatusOK, searchProducts)
	}
}

// user specific product
// accept two param age and gender if these param are not found it will take users
// age and gender to display the product
func GetUserSpecificProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		var user models.USer
		var age string
		var gender string
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		age = fmt.Sprintf("%d", c.Query("age"))
		gender = c.Query("gender")

		if age != "" && gender != "" {
			userId, exist := c.Get("uid")
			if !exist {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "User ID not found"})
				return
			}

			oid, err := primitive.ObjectIDFromHex(userId.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid user ID"})
				return
			}
			err = UserCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + err.Error()})
				return
			}
			// Calculate age
			currentDate := time.Now()
			dob := user.DOB
			ageCal := currentDate.Year() - dob.Year()

			// Adjust for the user's birthday not having occurred this year yet
			if currentDate.Month() < dob.Month() || (currentDate.Month() == dob.Month() && currentDate.Day() < dob.Day()) {
				ageCal--
			}
			age = fmt.Sprintf("%d", ageCal)
			gender = user.Gender
		}
		filter := bson.M{}

		if gender != "" {
			filter["gender"] = bson.M{
				"$in": bson.A{gender, "both"},
			}
		}

		filter["$expr"] = bson.M{
			"$and": bson.A{
				bson.M{"$gte": bson.A{
					bson.M{"$toInt": bson.M{"$split": bson.A{"$agegroup", "-"}[0]}},
					age,
				}},
				bson.M{"$lte": bson.A{
					bson.M{"$toInt": bson.M{"$split": bson.A{"$agegroup", "-"}[1]}},
					age,
				}},
			},
		}

		cursor, err := ProductCollection.Find(ctx, filter)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Something went wrong while fetching the data")
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &searchProducts); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "Invalid")
			return
		}

		// Iterate over each product and get pre-signed URLs for each image
		for i := range searchProducts {
			for j := range searchProducts[i].Image {
				// Get pre-signed URL for the image
				url, err := getPresignURL(searchProducts[i].Image[j])
				if err != nil {
					log.Println("Error generating pre-signed URL for image:", err)
					continue
				}
				// Update the image URL in the product
				searchProducts[i].Image[j] = url
			}
		}

		c.IndentedJSON(http.StatusOK, searchProducts)
	}
}
func ApproveProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check if the user is an admin
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		// Extract the product ID from the request parameters
		productID := c.Query("id")
		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid product ID"})
			return
		}

		// Find the product in the database
		var product models.Product
		err = ProductCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "product not found"})
			return
		}

		// Update the product as approved
		update := bson.M{"$set": bson.M{"approved": true, "isRejected": false}}
		_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "could not approve product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "product approved successfully", "product": product})
	}
}

func RejectProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check if the user is an admin
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		// Extract the product ID from the request parameters
		productID := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid product ID"})
			return
		}

		rejection_note := c.PostForm("rejection_note")

		// Find the product in the database
		var product models.Product

		// Update the product as rejected
		update := bson.M{"$set": bson.M{"approved": false, "isRejected": true, "rejection_note": rejection_note}}
		err = ProductCollection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update).Decode(&product)
		if err != nil {

			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"Error": "product not found"})
			}

			c.JSON(http.StatusBadRequest, gin.H{"Error": "could not reject product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "product rejected successfully", "product": product})
	}
}

func DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check admin permission
		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		// Extract product ID from URL parameter
		productID := c.Query("id")

		// Check if the provided product ID is valid
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Product ID can't be empty"})
			return
		}

		// Convert product ID to ObjectID
		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Product ID"})
			return
		}

		// Prepare filter to find product by ID
		filter := primitive.M{"_id": objID}

		//update and make it arhcive
		update := bson.M{"$set": bson.M{"isArchived": true, "approved": false, "isRejected": false}}

		//update
		result, err := ProductCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to delete product"})
			return
		}

		if result.MatchedCount < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to find Product"})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}

func FetchProductsAndReferencesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Aggregate pipeline to fetch products along with populated product references
		pipeline := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "ProductReferences",
					"localField":   "product_references.product_id",
					"foreignField": "_id",
					"as":           "product_references",
				},
			},
		}

		cursor, err := ProductCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var products []bson.M
		if err := cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProductReferenceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		pipeline := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "Products",
					"localField":   "product_id",
					"foreignField": "_id",
					"as":           "product",
				},
			},

			{
				"$lookup": bson.M{
					"from":         "seller",
					"localField":   "seller_id",
					"foreignField": "_id",
					"as":           "seller",
				},
			},
		}

		cursor, err := ProductReference.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var results []bson.M

		if err := cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

func getSuggestions(query string) ([]string, error) {
	// Simulated database or API call to fetch products and categories

	productsCursor, err := ProductCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var products []models.Product
	if err = productsCursor.All(context.Background(), &products); err != nil {
		return nil, err
	}

	categoriesCusror, err := CategoriesCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var categories []models.Categories
	if err = categoriesCusror.All(context.Background(), &categories); err != nil {
		return nil, err
	}

	var suggestions []string

	// Retrieve product suggestions
	for _, product := range products {
		if strings.Contains(strings.ToLower(product.Product_Name), strings.ToLower(query)) {
			suggestions = append(suggestions, product.Product_Name)
		}

	}

	// Retrieve category suggestions
	for _, category := range categories {
		if strings.Contains(strings.ToLower(category.Category), strings.ToLower(query)) {
			suggestions = append(suggestions, category.Category)
		}
	}

	return suggestions, nil
}

func SuggestionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		suggestions, err := getSuggestions(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if len(suggestions) > 10 {
			suggestions = suggestions[:10]
		}

		c.JSON(http.StatusOK, suggestions)
	}
}

//search product

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		query := c.Query("query")

		filter := bson.M{
			"$or": []bson.M{
				{"product_name": bson.M{"$regex": query, "$options": "i"}},

				{"category": bson.M{"$regex": query, "$options": "i"}},
			},
		}

		cursor, err := ProductCollection.Find(ctx, filter)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cursor.Close(ctx)

		var results []models.Product

		if err := cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//iterate through each product and add presignurl

		for i := range results {

			for j := range results[i].Image {

				url, err := getPresignURL(results[i].Image[j])

				if err != nil {

					log.Println("Error generating pre-signed URL for image:", err)
				}

				results[i].Image[j] = url

			}

		}

		c.JSON(http.StatusOK, results)

	}
}

func checkAdmin(ctx context.Context, c *gin.Context) bool {

	email, existse := c.Get("email")
	uid, exists := c.Get("uid")
	if !existse || !exists || uid == "" || email == "" {
		return false
	}
	uids := fmt.Sprintf("%v", uid)
	emails := fmt.Sprintf("%v", email)
	id, err := primitive.ObjectIDFromHex(uids)
	if err != nil {
		log.Fatal(err)

		return false
	}
	filter := primitive.M{
		"$and": []primitive.M{
			{"_id": id},
			{"email": emails},
		},
	}
	var seller models.Seller
	SellerCollection.FindOne(ctx, filter).Decode(&seller)
	if seller.User_type == utils.Admin {
		return true
	}
	return false
}

func checkSeller(ctx context.Context, c *gin.Context) bool {

	email, existse := c.Get("email")
	uid, exists := c.Get("uid")
	if !existse || !exists || uid == "" || email == "" {
		return false
	}
	uids := fmt.Sprintf("%v", uid)
	emails := fmt.Sprintf("%v", email)
	id, err := primitive.ObjectIDFromHex(uids)
	if err != nil {
		log.Fatal(err)

		return false
	}
	filter := primitive.M{
		"$and": []primitive.M{
			{"_id": id},
			{"email": emails},
		},
	}
	var seller models.Seller
	SellerCollection.FindOne(ctx, filter).Decode(&seller)
	if seller.User_type == utils.Seller {
		return true
	}
	return false
}

//handler to make a product featured

func MakeProductFeatured() gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		if !checkAdmin(ctx, c) {

			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		id := c.Param("id")

		isFeatured := c.Query("featured")

		if isFeatured != "true" && isFeatured != "false" {

			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid request"})

			return
		}

		isFeaturedBool, _ := strconv.ParseBool(isFeatured)

		objID, _ := primitive.ObjectIDFromHex(id)

		var product models.Product

		err := ProductCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to find product "})

			return

		}

		//get current time in RFC

		updateTime := time.Now().Format(time.RFC3339)

		update := bson.M{"$set": bson.M{"featured": isFeaturedBool, "updated_at": updateTime}}

		_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to update product"})

			return

		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})

	}
}

//get featured products

func GetFeaturedProducts() gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "updated_at", Value: -1}})
		findOptions.SetLimit(20)

		cursor, err := ProductCollection.Find(ctx, bson.M{"featured": true, "approved": true, "isRejected": false}, findOptions)

		if err != nil {

			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong while fetching the data")

			return

		}

		var featuredProducts []models.Product

		err = cursor.All(ctx, &featuredProducts)

		if err != nil {

			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong while fetching the data")

			return

		}
		//append prsign url for each imabge of each product

		for i, product := range featuredProducts {

			for j, image := range product.Image {

				url, err := getPresignURL(image)

				if err != nil {

					log.Println("Error generating pre-signed URL for image:", err)

					continue

				}

				featuredProducts[i].Image[j] = url

			}

		}

		c.JSON(http.StatusOK, featuredProducts)

	}

}
func GetSellerProductForAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		var sellerId = c.Param("id")

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

func GetProductsCsv(c *gin.Context) {
	var searchProducts []models.Product

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := ProductCollection.Find(ctx, bson.M{})
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Error fetching products: " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &searchProducts); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error decoding products: " + err.Error()})
		return
	}

	// Pre-sign image URLs
	for i := range searchProducts {
		for j := range searchProducts[i].Image {
			url, err := getPresignURL(searchProducts[i].Image[j])
			if err != nil {
				log.Println("Error generating pre-signed URL for image:", err)
				continue
			}
			searchProducts[i].Image[j] = url
		}
	}

	// Prepare CSV headers
	headers := []string{"Product Name", "Category", "Price", "Images", "Description", "isApproved", "isFeatured", "isRejected", "Price Range"}

	// Prepare CSV rows
	var rows [][]string
	for _, product := range searchProducts {
		priceFloat, err := strconv.ParseFloat(product.Price, 64)
		if err != nil {
			priceFloat = 0.0
		}

		row := []string{
			product.Product_Name,
			product.Category,
			strconv.FormatFloat(priceFloat, 'f', 2, 64),
			strings.Join(product.Image, ";"),
			product.Discription,
			strconv.FormatBool(product.Approved),
			strconv.FormatBool(product.Featured),
			strconv.FormatBool(product.IsRejected),
		}

		var priceRange []string
		for _, price := range product.PriceRange {
			priceRange = append(priceRange, "["+strconv.FormatFloat(float64(price.MinQuantity), 'f', 2, 64)+"-"+strconv.FormatFloat(float64(price.MaxQuantity), 'f', 2, 64)+":"+price.Price+"]")
		}
		row = append(row, strings.Join(priceRange, ";"))

		rows = append(rows, row)
	}

	// Generate CSV data
	csvData, err := GenerateCSV(headers, rows)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set CSV-specific headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=products.csv")
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Transfer-Encoding", "binary")

	// Write the CSV data to the response
	c.Data(http.StatusOK, "text/csv", csvData)
}
