package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileHeader.Filename),
		Body:   fileReader,
	})
	if err != nil {
		return "", err
	}

	// Get the URL of the uploaded file
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, fileHeader.Filename)

	return url, nil
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var product models.Product
		defer cancel()
		// if err := c.BindJSON(&product); err != nil {
		// 	log.Println("error while binding")
		// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		// 	return
		// 
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		product.Product_ID = primitive.NewObjectID()
		product.Product_Name = c.PostForm("product_name")
		//product.SKU = c.PostForm("sku")
		count, err := ProductCollection.CountDocuments(ctx, primitive.M{"product_name": product.Product_Name})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if count > 0 {
			log.Println("Error")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "product with this name is already present"})
			return
		}

		//	log.Println(strings.TrimSpace(c.PostForm("price")))
		price := strings.TrimSpace(c.PostForm("price"))
		if err != nil {
			log.Println("error while parsing price")
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		product.Price = price

		product.Discription = c.PostForm("discription")
		product.Category = c.PostForm("category")
		//handling file
		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		files := form.File["files"]
		fmt.Println("Fetched filesa:")
		fmt.Println(files)
		//var fileString []string
		cfg, err := config.LoadDefaultConfig(context.TODO())
		fmt.Println("region", cfg.Region)
		if err != nil {
			log.Fatal(err)
			log.Println("error while multipart")
			c.String(http.StatusInternalServerError, "get form err: %s", err.Error())
			return
		}
		//	client := s3.NewFromConfig(cfg)
		//	uploader := manager.NewUploader(client)
		var errors []string
		var uploadedURLs []string
		var resFileName []string
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
		var product models.Product
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
		err = ProductCollection.FindOne(ctx, primitive.M{"_id": prodID}).Decode(&product)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, utils.ErrorCantFindProduct)
			return
		}
		c.JSON(http.StatusOK, product)
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
		var SearchProducts []models.Product
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
			c.IndentedJSON(404, "Something went wrong while fetching the data")
			return
		}

		err = cursor.All(ctx, &SearchProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid")
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid request")
			return
		}
		c.IndentedJSON(200, SearchProducts)

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

