package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gin-gonic/gin"
)

func main() {
	client, _ := clerk.NewClient("sk_test_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	// Use Gin engine
	router := gin.Default()

	// Middleware to verify Session Token from Authorization header
	verifyToken := func(c *gin.Context) {
		sessionToken := c.Request.Header.Get("Authorization")
		if sessionToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		sessionToken = strings.TrimPrefix(sessionToken, "Bearer ")
		sessClaims, err := client.VerifyToken(sessionToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
			return
		}

		fmt.Println(sessClaims.Claims)
		// Store user info in context for later access
		c.Set("user", sessClaims.Claims.Subject)
	}

	// Protected route for saying hello
	router.GET("/albums", verifyToken, func(c *gin.Context) {
		// check if this makes a new request to the server or is it done on server
		user, err := client.Users().Read(c.MustGet("user").(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Welcome, %s!", *user.FirstName)})
	})

	// Run the server
	router.Run(":8080")
}
