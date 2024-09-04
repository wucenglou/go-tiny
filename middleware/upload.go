package middleware

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadMiddleware is a middleware that handles file uploads.
func UploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from the form data
		file, err := c.FormFile("avatar")
		if err != nil {
			c.JSON(400, gin.H{"error": "No file was uploaded"})
			c.Abort()
			return
		}

		// Check the file extension
		ext := filepath.Ext(file.Filename)
		validExtensions := []string{".jpg", ".jpeg", ".png", ".gif"}
		if ext == "" || !isValidExtension(ext, validExtensions) {
			c.JSON(400, gin.H{"error": "Invalid file format"})
			c.Abort()
			return
		}

		// Generate a unique filename
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// Define the upload directory
		uploadDir := "./uploads/"
		err = os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create directory"})
			c.Abort()
			return
		}

		// Open the file
		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to open file"})
			c.Abort()
			return
		}
		defer src.Close()

		// Save the file to disk
		dst, err := os.Create(filepath.Join(uploadDir, filename))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to save file"})
			c.Abort()
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			c.JSON(500, gin.H{"error": "Failed to copy file"})
			c.Abort()
			return
		}

		// Pass the filename to the next handler
		c.Set("filename", filename)
		c.Next()
	}
}

// isValidExtension checks if the file extension is valid.
func isValidExtension(ext string, validExts []string) bool {
	for _, v := range validExts {
		if ext == v {
			return true
		}
	}
	return false
}
