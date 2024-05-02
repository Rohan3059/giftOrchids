package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var EnquireCollection *mongo.Collection = database.ProductData(database.Client, "enquire")
var RequirementMessageCollection *mongo.Collection = database.ProductData(database.Client,"RequirementMessage")
var ReviewsCollection *mongo.Collection = database.ProductData(database.Client, "Reviews")
var FeedsCollection *mongo.Collection = database.ProductData(database.Client, "New_Feeds")



var uploader *s3manager.Uploader






func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read AWS credentials from environment variables
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

	// Create an uploader instance using the AWS session
	uploader = s3manager.NewUploader(awsSession)
}



func saveFile(fileReader io.Reader, fileHeader *multipart.FileHeader) (string, error) {
	// Upload the file to S3 using the fileReader
	
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	mtype, error := mimetype.DetectReader(fileReader);
	if error != nil {
		return "", error
	}

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileHeader.Filename),
		Body:   fileReader,
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

func getPresignURL(s3Url string) (string, error) {
	// Create an S3 service client using the provided session

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read AWS credentials from environment variables
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

	// Generate the pre-signed URL
	req, _ := s3client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
		
	})
	presignedURL, err := req.Presign(time.Hour * 24) // URL expires in 24 hours
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}



func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var sellerId string
		if(checkSeller(ctx, c)){
			sellerID, exists := c.Get("uid")
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Seller ID not found in context"})
				return
			}

			sellerId = sellerID.(string)
			
		}

		var product models.Product
		
		defer cancel()
		
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		product.Product_ID = primitive.NewObjectID()
		product.Product_Name = c.PostForm("product_name")

		product.SKU = c.PostForm("sku")
		attributesList := c.PostForm("attributes")
		

		var attributes []models.AttributeValue
		if err := json.Unmarshal([]byte(attributesList), &attributes); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		product.Attributes = attributes

		
		count, err := ProductCollection.CountDocuments(ctx, primitive.M{"sku": product.SKU})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if count > 0 {
			log.Println("Error")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "product with this name is already present"})
			return
		}

		
		price := strings.TrimSpace(c.PostForm("price"))
		if err != nil {
			log.Println("error while parsing price")
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		product.Price = price

		product.Discription = c.PostForm("discription")
		product.Category = c.PostForm("category")
		
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

		var sellers[]string
		sellers = append(sellers, sellerId)
		product.SellerRegistered = sellers

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
	c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()} )
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if !checkAdmin(ctx, c) {
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
		newProduct, err := bson.Marshal(product)
		if err != nil {
			log.Fatal(err)
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}

		update := bson.M{"$set": bson.Raw(newProduct)}
		result,err := ProductCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			c.Abort()
			return
		}
		if result.ModifiedCount < 1 {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to find Product "})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, product)
	}

}

// accept diff filter to get product
func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		queryParam := c.Query("name")
		filter := []primitive.M{}
		if queryParam != "" {
			filter = append(filter, primitive.M{"$regex": queryParam})
		}
		productCategory := c.Query("category")
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
		finalFilter := primitive.M{}
		if len(filter) > 0 {
			finalFilter = primitive.M{"$and": filter}
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, finalFilter)
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
				url, err := getPresignURL( searchProducts[i].Image[j])
				if err != nil {
					log.Println("Error generating pre-signed URL for image:", err)
					continue
				}
				// Update the image URL in the product
				searchProducts[i].Image[j]= url
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
		fmt.Println("ProductID=="+productID);
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
		update := bson.M{"$set": bson.M{"approved": true}}
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

		// Find the product in the database
		var product models.Product
		err = ProductCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"Error": "product not found"})
			return
		}

		// Update the product as rejected
		update := bson.M{"$set": bson.M{"approved": false}}
		_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "could not reject product"})
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

		// Perform delete operation
		result, err := ProductCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
			return
		}

		// Check if any document was deleted
		if result.DeletedCount < 1 {
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

func SuggestionsHandler() gin.HandlerFunc  {
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

		c.JSON(http.StatusOK, results)
	
	}
}



func AddReviewHandler()gin.HandlerFunc { return func(c *gin.Context){
	// Parse request body to get review details
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	var review models.Reviews
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId,exist := c.Get("uid")
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

	
	if(review.ReviewsDetails.ReviewRating > 5 || review.ReviewsDetails.ReviewRating < 1) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	if(review.ReviewsDetails.ReviewText == "" || review.ReviewsDetails.ReviewTitle == "") {
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

	// Update product with new review ID
	// Append new review ID to existing array of review IDs in Product collection
	update := bson.M{"$push": bson.M{"reviews": review.Id}}
	_, err = ProductCollection.UpdateOne(ctx, bson.M{"_id": review.ProductId}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product with review ID"})
		return
	}


	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Thanks for your review!"})
}
}

func ApproveReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if !checkAdmin( ctx,c){

			c.JSON(http.StatusForbidden, gin.H{"Error": "You're not authorized to approve reviews"})
			return

		}		

		var review models.Reviews
		if err := c.ShouldBindJSON(&review); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if review.Id.Hex() == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Review ID is required"})
			return
		}


		filter := bson.M{"_id": review.Id}
		update := bson.M{"$set": bson.M{"approved": true}}
		_, err := ReviewsCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to approve review"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Review approved successfully"})


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
                    "path": "$user",
                    "preserveNullAndEmptyArrays": true,
                },
            },
			 {
        "$project": bson.M{
			"_id" : 1,
			"product_id" : 1,
			"reviews_details":1,
			"approved" : 1,
			"archived":1,
			"created_at":1,
			
			"user" : bson.M{
            "_id":    "$user._id",
            "name":   "$user.username",
            "mobileno":  "$user.mobileno",
            
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
    "ratingPercentage":	percentageReviews ,
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
                    "path": "$user",
                    "preserveNullAndEmptyArrays": true,
                },
            },
			 {
        "$project": bson.M{
			"_id" : 1,
			"product_id" : 1,
			"reviews_details":1,
			"approved" : 1,
			"archived":1,
			"created_at":1,
			
			"user" : bson.M{
            "_id":    "$user._id",
            "name":   "$user.username",
            "mobileno":  "$user.mobileno",
            
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
                    "path": "$user",
                    "preserveNullAndEmptyArrays": true,
                },
				
            },{
				"$unwind": bson.M{
					"path": "$product",
					"preserveNullAndEmptyArrays": true,
				},
			},
			 {
        "$project": bson.M{
			"_id" : 1,
			"product" : 1,
			"reviews_details":1,
			"approved" : 1,
			"archived":1,
			"created_at":1,
			
			"user" : bson.M{
            "_id":    "$user._id",
            "name":   "$user.username",
            "mobileno":  "$user.mobileno",
            
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

