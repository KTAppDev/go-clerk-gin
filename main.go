package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	key := os.Getenv("CLERK_SECRET_KEY")
	client, _ := clerk.NewClient(key)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Middleware to verify Session Token from Authorization header
	verifyToken := func(c *gin.Context) {
		sessionToken := c.Request.Header.Get("Authorization")
		if sessionToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// NOTE: Get cookies
		//  cookieToken, _ := c.Request.Cookie("__session")
		// clientUat, _ := c.Request.Cookie("__client_uat")

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
	// router.GEt takes three items: path middleware and handler
	router.GET("/protected", verifyToken, func(c *gin.Context) {
		// check if this makes a new request to the server or is it done on server
		email := client.Emails()
		fmt.Println(email) // print out emails)
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
