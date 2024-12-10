package main

import (
	"github.com/gin-gonic/gin"
)

func getReceiptById(c *gin.Context) {
	// GET: /receipts/{id}/points
}

func processReceipts(c *gin.Context) {
	// POST: /receipts/process
}

func main() {
	router := gin.Default()
	router.GET("/receipts/:id/points", getReceiptById)
	router.POST("/receipts/process", processReceipts)

	router.Run("localhost:8080")
}
