// main.go

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

// main is the entry point of the application
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Get Clerk secret key from environment variable
	key := os.Getenv("CLERK_SECRET_KEY")

	// Create a new Clerk client instance
	client, _ := clerk.NewClient(key)

	// Create a new Gin router instance
	router := gin.Default()

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// VerifyToken is a middleware function that verifies the session token
	// from the Authorization header
	verifyToken := func(c *gin.Context) {
		sessionToken := c.Request.Header.Get("Authorization")
		if sessionToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Remove the "Bearer " prefix from the session token
		sessionToken = strings.TrimPrefix(sessionToken, "Bearer ")

		// Verify the session token using the Clerk client
		sessClaims, err := client.VerifyToken(sessionToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session token"})
			return
		}

		// Print the session claims
		fmt.Println(sessClaims.Claims)

		// Store the user info in the context for later access
		c.Set("user", sessClaims.Claims.Subject)
	}

	// Protected is a handler function that returns a welcome message
	// to the authenticated user
	protected := func(c *gin.Context) {
		// Get the user info from the context
		userID := c.MustGet("user").(string)

		// Get the user's email addresses using the Clerk client
		email := client.Emails()

		fmt.Println(email) // print out emails

		// Get the user's profile using the Clerk client
		user, err := client.Users().Read(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
			return
		}

		// Return a welcome message to the user
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Welcome, %s!", *user.FirstName)})
	}

	// Register the protected route
	router.GET("/protected", verifyToken, protected)

	// Run the server
	router.Run(":8080")
}