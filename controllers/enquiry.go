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
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func EnquiryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		mobileno, existse := c.Get("mobile")

		uid, exists := c.Get("uid")

		if !existse || !exists || uid == "" || mobileno == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		var enquire models.Enquire
		if err := c.BindJSON(&enquire); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err})
			return

		}

		if enquire.Enquiry_note == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Enquiry note is required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		enquire.Enquire_id = primitive.NewObjectID()
		enquire.User_id = uid.(string)

		enquire.EnquireId = strconv.FormatInt(time.Now().Unix(), 10)[2:10]
		enquire.Status = "Pending"
		enquire.Enquire_date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		enquire.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		_, err := EnquireCollection.InsertOne(ctx, enquire)
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Success": "enquiry registerd"})
	}
}

// update status of requirement
func UpdateEnquireStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "forbidden"})
			return
		}

		id := c.Param("id")
		status := c.Query("status")
		//objectid

		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		//find  update enquiry status
		var enquiry models.Enquire

		filter := bson.M{"_id": objectId}

		update := bson.M{"$set": bson.M{"status": status}}

		findErr := EnquireCollection.FindOneAndUpdate(ctx, filter, update).Decode(&enquiry)

		if findErr != nil {
			if errors.Is(findErr, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"Error": "Enquiry not found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"Error": "something went wrong"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": "Enquiry status updated"})

	}
}

func GetUserEnquiries() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from token
		uid, exists := c.Get("uid")
		if !exists || uid == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// Convert user ID to string
		userID := uid.(string)
		fmt.Println(userID)
		// Define filter to fetch enquiries for the specific user
		filter := bson.M{"user_id": userID}

		// Fetch enquiries from the database
		var enquiries []map[string]interface{}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := EnquireCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var enquire models.Enquire
			if err := cursor.Decode(&enquire); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
				return
			}

			// Fetch product details based on product_id
			var product models.Product
			fmt.Println(enquire)

			prodID, err := primitive.ObjectIDFromHex(enquire.Product_id)
			if err != nil {
				c.Header("content-type", "application/json")
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
				c.Abort()
				return
			}

			errors := ProductCollection.FindOne(ctx, bson.M{"_id": prodID}).Decode(&product)
			if errors != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to fetch product details"})
				return
			}

			var imgUrl string
			if product.Image != nil {
				url, err := getPresignURL(product.Image[0])
				if err != nil {
					imgUrl = ""
				}
				imgUrl = url
			}

			// Add product name and image to enquiry
			enquiryWithProduct := map[string]interface{}{
				"enquiry_id":       enquire.Enquire_id.Hex(),
				"user_id":          enquire.User_id,
				"product_name":     product.Product_Name,
				"product_image":    imgUrl,
				"enquire_note":     enquire.Enquiry_note,
				"enquire_quantity": enquire.Quantity,
				"enquire_status":   enquire.Status,
				// Add other enquiry fields if needed
			}
			enquiries = append(enquiries, enquiryWithProduct)
		}
		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "something went wrong"})
			return
		}

		c.JSON(http.StatusOK, enquiries)
	}
}

func GETEnquiryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "Admin Token Not found"})
			return
		}

		var enquire []models.Enquire

		cursor, err := EnquireCollection.Find(ctx, primitive.M{})
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch"})
			return
		}
		if err := cursor.All(ctx, &enquire); err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		defer cursor.Close(ctx)

		// Enrich enquiry data with additional details
		enquiriesWithDetails := make([]map[string]interface{}, 0)

		for _, enquiry := range enquire {
			// Fetch product details based on product_id
			productDetails := getProductDetails(ctx, enquiry.Product_id)

			// Fetch user details based on user_id
			userDetails := getUserDetails(ctx, enquiry.User_id)

			// Construct enriched enquiry
			enquiryWithDetails := map[string]interface{}{
				"enquiry": enquiry,
				"product": productDetails,
				"user":    userDetails,
			}

			enquiriesWithDetails = append(enquiriesWithDetails, enquiryWithDetails)
		}

		c.JSON(http.StatusOK, enquiriesWithDetails)
	}
}

func GetAdminSingleEnquiry() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var Enquire_id = c.Param("id")
		id, err := primitive.ObjectIDFromHex(Enquire_id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ID"})
			return
		}

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusForbidden, gin.H{"Error": "Admin Token Not found"})
			return
		}

		var enquire models.Enquire

		findErr := EnquireCollection.FindOne(ctx, primitive.M{
			"_id": id,
		}).Decode(&enquire)
		if findErr != nil {
			log.Print(findErr)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to fetch"})
			return
		}

		// Fetch product details based on product_id
		productDetails := getProductDetails(ctx, enquire.Product_id)

		// Fetch user details based on user_id
		userDetails := getUserDetails(ctx, enquire.User_id)

		// Construct enriched enquiry
		enquiryWithDetails := map[string]interface{}{
			"enquiry": enquire,
			"product": productDetails,
			"user":    userDetails,
		}

		c.JSON(http.StatusOK, enquiryWithDetails)
	}
}

// Function to fetch product details based on product_id
func getProductDetails(ctx context.Context, productID string) map[string]interface{} {
	var productDetails models.Product

	id, err := primitive.ObjectIDFromHex(productID)

	if err != nil {
		log.Printf("Error parsing product ID %s: %s", productID, err.Error())
		return nil
	}

	errs := ProductCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&productDetails)
	if errs != nil {
		log.Printf("Error fetching product details for product ID %s: %s", productID, errs.Error())
		return nil
	}

	for i, url := range productDetails.Image {

		// Get pre-signed URL for the image
		url, err := getPresignURL(url)
		if err != nil {
			log.Println("Error generating pre-signed URL for image:", err)
			continue
		}
		// Update the image URL in the product
		productDetails.Image[i] = url

	}

	var sellerArray []map[string]interface{}
	for _, seller := range productDetails.SellerRegistered {
		//get seller details
		sellerDetail := getSellerDetails(ctx, seller)
		sellerArray = append(sellerArray, sellerDetail)
	}

	newProductDetails := map[string]interface{}{
		"name":        productDetails.Product_Name,
		"_id":         productDetails.Product_ID,
		"image":       productDetails.Image,
		"price":       productDetails.Price,
		"category":    productDetails.Category,
		"price_range": productDetails.PriceRange,
		"sellers":     sellerArray,
	}

	return newProductDetails
}

// Function to fetch user details based on user_id
func getUserDetails(ctx context.Context, userID string) map[string]interface{} {
	var userDetails models.USer

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("Error parsing user ID %s: %s", userID, err.Error())
		return nil
	}

	errs := UserCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&userDetails)
	if errs != nil {
		log.Printf("Error fetching user details for user ID %s: %s", userID, errs.Error())
		return nil
	}

	//only send name, email,mobile and _id not other details , just create new map for it
	newUserDetails := map[string]interface{}{
		"name":   userDetails.UserName,
		"email":  userDetails.Email,
		"mobile": userDetails.MobileNo,
		"_id":    userDetails.User_id,
	}

	return newUserDetails
}

func getSellerDetails(ctx context.Context, id string) map[string]interface{} {

	var sellerDetails models.Seller

	sellerId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Error parsing seller ID %s: %s", id, err.Error())
		return nil
	}

	errs := SellerCollection.FindOne(ctx, bson.M{"_id": sellerId}).Decode(&sellerDetails)
	if errs != nil {

		log.Printf("Error fetching seller details for seller ID %s: %s", id, errs.Error())
		return nil

	}

	//only send name, email,mobile and _id not other details , just create new map for it
	newSellerDetails := map[string]interface{}{
		"company_name": sellerDetails.Company_Name,
		"email":        sellerDetails.Email,
		"mobile":       sellerDetails.MobileNo,
		"_id":          sellerDetails.Seller_ID,
	}

	return newSellerDetails

}

// GetAllRequirementMessages retrieves all RequirementMessages
func GetAllRequirementMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cursor, err := RequirementMessageCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var messages []models.RequirementMessage
		if err := cursor.All(ctx, &messages); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}

// CreateRequirementMessage creates a new RequirementMessage
func CreateRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate new ObjectID
		message.Requirement_id = primitive.NewObjectID()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := RequirementMessageCollection.InsertOne(ctx, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, result.InsertedID)
	}
}

// UpdateRequirementMessage updates a RequirementMessage
func UpdateRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": message.Requirement_id}
		update := bson.M{"$set": message}

		_, err := RequirementMessageCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "RequirementMessage updated successfully"})
	}
}

// DeleteRequirementMessage deletes a RequirementMessage
func DeleteRequirementMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var message models.RequirementMessage
		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": message.Requirement_id}

		_, err := RequirementMessageCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "RequirementMessage deleted successfully"})
	}
}
