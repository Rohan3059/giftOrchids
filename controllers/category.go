package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var CategoriesCollection *mongo.Collection = database.ProductData(database.Client, "Categories")
type CategoryWithChildren struct {
    Category       models.Categories
    ChildCategories []models.Categories
}
func AddCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var category models.Categories
		
		category.Category_ID = primitive.NewObjectID()

		categoryName := c.PostForm("category")

		parentCategoryId := c.PostForm("parent_category")

		

		if categoryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "category name is required"})
			return
		}

		if parentCategoryId!=""  {
			category.Parent_Category, _ = primitive.ObjectIDFromHex(parentCategoryId)
		}
		category.Category = categoryName

		form, err := c.MultipartForm()
		if err != nil {
			log.Println("error while multipart")
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			
		}

		image := form.File["image"]
		if len(image) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "image is required"})
			return
		}

		categoryImageHeader,err := image[0].Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		defer categoryImageHeader.Close()
		
		categoryImage, err := saveFile(categoryImageHeader, image[0])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		category.Category_image = categoryImage

		category.Category_Description = c.PostForm("category_description")

		category.Approved = false
		
		count, err := CategoriesCollection.CountDocuments(ctx, bson.M{"category": category.Category})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Category already exist with this name"})
			return
		}
	
		_, anyerr := CategoriesCollection.InsertOne(ctx, category)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Category Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Succesfully added category")

	}
}
func GetCategory() gin.HandlerFunc {
    return func(c *gin.Context) {
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Aggregation pipeline to perform $lookup with parent_category collection
        pipeline :=  []bson.M{
                {
					"$lookup": bson.M{
                    "from":         "Categories", 
                    "localField":   "parent_category",
                    "foreignField": "_id",
                    "as":           "parent_category_details",
                },
            },
		}
           
        

        // Execute aggregation pipeline
        cursor, err := CategoriesCollection.Aggregate(ctx, pipeline)
        if err != nil {
            c.JSON(http.StatusInternalServerError, "Something went wrong. Please try again.")
            return
        }
        defer cursor.Close(ctx)

		var results []models.Categories;

        // Decode the results into category slice
        if err := cursor.All(ctx, &results); err != nil {
            c.JSON(http.StatusInternalServerError, "Something went wrong while fetching data. Please try again.")
            return
        }

        // Loop through the cursor and get image of each category
        for i := range results {
            url, err := getPresignURL(results[i].Category_image)
            if err != nil {
                log.Println("Error generating pre-signed URL for image:", err)
                continue
            }
            if url != "" {
                results[i].Category_image = url
            }
        }

        c.JSON(http.StatusOK, results)
    }
}




func GetSingleCategory() gin.HandlerFunc {
    // Extract category ID from query parameter
	return func(c *gin.Context) {
	categoryID := c.Query("id")
    if categoryID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"Error": "Category ID is required"})
        return
    }

    // Convert category ID string to ObjectID
    objID, err := primitive.ObjectIDFromHex(categoryID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid category ID"})
        return
    }

	// Find category by ID
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var category models.Categories
	err = CategoriesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Main category not found"})
		return
	}

  

	url, err := getPresignURL(category.Category_image)
	if err != nil {
		log.Println("Error generating pre-signed URL for image:", err)
		url = ""
	}
	if url != "" {
		category.Category_image = url
	}


	
    
    child_category, err := GetCategoryWithId(objID)
	


	 categoryWithChildren := CategoryWithChildren{
            Category:       category,
            ChildCategories: child_category,
        }


    c.JSON(http.StatusOK,categoryWithChildren )
	}
}

func GetCategoryWithId(categoryID primitive.ObjectID) ([]models.Categories, error) {
    var ctx = context.Background()

    // Aggregation pipeline to find category details and its child categories recursively
    pipeline := []bson.M{
        {
            "$match": bson.M{"parent_category": categoryID},
        },
        {
            "$graphLookup": bson.M{
                "from":              "Categories",
                "startWith":         "$_id",
                "connectFromField":  "parent_category",
                "connectToField":    "_id",
                "as":                "child_categories",
                "maxDepth":          10, // Set a maximum depth to prevent infinite recursion
            },
        },
    }

    // Execute aggregation pipeline
    cursor, err := CategoriesCollection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    // Decode the results into a slice of categories
    var categories []models.Categories
    for cursor.Next(ctx) {
        var category models.Categories
        if err := cursor.Decode(&category); err != nil {
            return nil, err
        }
		// Get image of each category prsign url

		url, err := getPresignURL(category.Category_image)
		if err != nil {
			log.Println("Error generating pre-signed URL for image:", err)
			url = ""
		}
		if url != "" {
			category.Category_image = url
		}

        categories = append(categories, category)
    }

    // Check if any categories found
    if len(categories) == 0 {
        return nil, errors.New("no categories found")
    }

    return categories, nil
}



func EditCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cat_id := c.Query("cat_id")

		if cat_id == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid user id"})
			c.Abort()
			return
		}

		catID, err := primitive.ObjectIDFromHex(cat_id)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
			return
		}

		var Editcategory models.Categories
		if err := c.BindJSON(&Editcategory); err != nil {
			c.IndentedJSON(http.StatusBadRequest, "category is not in correct format")
			return
		}
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: catID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "category", Value: Editcategory.Category}}}}
		_, err = CategoriesCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully updated category")

	}
}


func ApproveCategory() gin.HandlerFunc {


	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cat_id := c.Query("id")

		if cat_id == "" {
			c.Header("content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Category id"})
			c.Abort()
			return
		}

		catID, err := primitive.ObjectIDFromHex(cat_id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
			return
		}

		
		_,err = CategoriesCollection.UpdateOne(ctx, bson.M{"_id": catID}, bson.D{{Key: "$set", Value: bson.M{"isApproved": true}}})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category approved successfully"})
		

	}


}