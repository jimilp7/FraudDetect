package api

import (
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"os"
	"path/filepath"
)

// uploadTransactions godoc
// @Summary Upload transactions
// @Description This endpoint is used to upload a CSV file containing transaction data.
// @Tags transactions
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Transaction file"
// @Success 200 {object} map[string]string "file_id"
// @Router /upload [post]
func uploadTransactions(c *gin.Context) {
	// Accept a file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Generate a NanoID
	id, err := gonanoid.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate an ID"})
		return
	}

	// Ensure TransactionFiles directory exists
	err = os.MkdirAll("TransactionFiles", os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Define the path to save the file
	filepath := filepath.Join("TransactionFiles", id+".csv")

	// Save the file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	// Respond with the NanoID
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "file_id": id})
}
