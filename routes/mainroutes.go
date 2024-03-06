package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


func AdminRoutes(adminRoutes *gin.Engine) {
	
	/*adminRoutes.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"PUT", "GET", "POST", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "token", "Content-Type","Token"},
			ExposeHeaders:    []string{"Content-Length", "Content-Type","token"},
			
			AllowCredentials: true,
		}))
	*/
	configs := cors.DefaultConfig()
    configs.AllowAllOrigins = true
    configs.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    configs.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Token","token"} // Add "Token" header

}
