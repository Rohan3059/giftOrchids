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
		port = "8080"
	}

	router = gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)

	log.Fatal(router.Run(":" + port))
	router.GET("/", controllers.GetCategory())
	router.GET("/getproduct", controllers.ValidateOtpHandler())

}
