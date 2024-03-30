package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		
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
		
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "No authorization header is provided"})
			c.Abort()
			return
		}

		calims, msg := tokens.ValidateUSERToken(clientToken)
	
		if msg != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			c.Abort()
			return

		}
		
	
		c.Set("mobile", calims.MobileNo)
		c.Set("uid", calims.Uid)
		c.Next()

	}
}
