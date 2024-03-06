package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Print("checking admin token")
		clientToken := c.Request.Header.Get("token")
			fmt.Print(clientToken)
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "No authorization header is provided"})
			c.Abort()
			return
		}

		calims, msg := tokens.ValidateToken(clientToken)

		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			c.Abort()
			return

		}

		c.Set("email", calims.Email)
		c.Set("uid", calims.Uid)
		c.Next()

	}
}

func UserAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		fmt.Print("Checking Header token in user authentication\n")
		clientToken := c.Request.Header.Get("token")
		fmt.Print(clientToken)
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "No authorization header is provided"})
			c.Abort()
			return
		}

		calims, msg := tokens.ValidateUSERToken(clientToken)
		fmt.Print(calims)
		fmt.Print("CHcking calims\n")

		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			c.Abort()
			return

		}
		
		fmt.Print("Checking Message in user authentication\n")

		c.Set("mobile", calims.MobileNo)
		c.Set("uid", calims.Uid)
		c.Next()

	}
}
