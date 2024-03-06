package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohan3059/bizGrowth/database"
	"github.com/rohan3059/bizGrowth/models"
	"github.com/rohan3059/bizGrowth/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var EnquireCollection *mongo.Collection = database.ProductData(database.Client, "enquire")

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var product models.Product
		defer cancel()
		// if err := c.BindJSON(&product); err != nil {
		// 	log.Println("error while binding")
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }
		product.Product_ID = primitive.NewObjectID()
		product.Product_Name = c.PostForm("product_name")
		//product.SKU = c.PostForm("sku")
		count, err := ProductCollection.CountDocuments(ctx, primitive.M{"product_name": product.Product_Name})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if count > 0 {
			log.Println("error")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "product with this name is already present"})
			return
		}

		log.Println(strings.TrimSpace(c.PostForm("price")))
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
		fmt.Println(files)
		var fileString []string
		for _, file := range files {
			filename := filepath.Base(file.Filename)
			fileString = append(fileString, filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
				return
			}
		}
		fmt.Println(fileString)
		product.Image = fileString

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
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		count, err := ProductCollection.CountDocuments(ctx, primitive.M{"sku": product.SKU})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			log.Println("error")
			c.JSON(http.StatusBadRequest, gin.H{"error": "product with this SKU is already present"})
			return
		}
		if product.Product_ID.Hex() == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
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
		ProductCollection.UpdateOne(ctx, filter, update)

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
