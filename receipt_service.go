package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type receiptItem struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type receipt struct {
	// Id           string        `json:"id"`
	Retailer     string        `json:"retailer"`
	PurchaseDate string        `json:"purchaseDate"`
	PurchaseTime string        `json:"purchaseTime"`
	Items        []receiptItem `json:"items"`
	Total        string        `json:"total"`
}

type receiptStore struct {
	receipts map[string]receipt
}

func (store *receiptStore) getReceiptById(c *gin.Context) {
	// GET: /receipts/{id}/points
	id := c.Param("id")

	receiptGet, ok := store.receipts[id]
	if !ok {
		// TODO: handle error
		return
	}

	c.IndentedJSON(http.StatusOK, receiptGet)
}

func (store *receiptStore) processReceipts(c *gin.Context) {
	// POST: /receipts/process

	var newReceipt receipt
	if err := c.BindJSON(&newReceipt); err != nil {
		// TODO: handle error
		return
	}

	// assumes duplicate receipts with identical content are allowed
	id := uuid.New().String()
	store.receipts[id] = newReceipt

	returnIdJSON := struct {
		Id string `json:"id"`
	}{Id: id}

	c.IndentedJSON(http.StatusOK, returnIdJSON)
}

func setupRouter(store *receiptStore) *gin.Engine {
	router := gin.Default()
	router.GET("/receipts/:id/points", store.getReceiptById)
	router.POST("/receipts/process", store.processReceipts)
	return router
}
