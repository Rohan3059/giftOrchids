package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/controllers"
	"github.com/kravi0/BizGrowth-backend/routes"
	cors "github.com/rs/cors/wrapper/gin"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	//app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router = gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	//routes.AdminRoutes(router)
	// router.Use(middleware.Authentication())
	// router.GET("/addtocart", app.AddtoCart())
	// router.GET("/removeitem", app.RemoveItem())
	// router.GET("/listcart", controllers.GetItemFromCart())
	// router.POST("/addaddress", controllers.AddAddress())
	// router.PUT("/edithomeaddress", controllers.EditHomeAddress())
	// router.PUT("/editworkaddress", controllers.EditWorkAddress())
	// router.GET("/deleteaddresses", controllers.DeleteAddress())
	// router.GET("/cartcheckout", app.BuyFromCart())
	// router.GET("/instantbuy", app.InstantBuy())
	log.Fatal(router.Run(":" + port))
	router.GET("/", controllers.GetCategory())
	router.GET("/getproduct", controllers.ValidateOtpHandler())

}
