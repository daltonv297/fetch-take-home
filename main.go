package main

import "github.com/gin-gonic/gin"

func main() {
	store := &receiptStore{
		receipts: make(map[string]receiptWrapper),
	}
	var router *gin.Engine = setupRouter(store)

	router.Run("localhost:8080")
}
