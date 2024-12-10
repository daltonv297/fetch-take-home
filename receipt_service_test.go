package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() *gin.Engine {
	store := &receiptStore{
		receipts: make(map[string]receipt),
	}
	router := setupRouter(store)
	return router
}

func TestProcessAndReceive(t *testing.T) {
	router := setupTestContext()

	testReceipt := receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []receiptItem{
			{
				ShortDescription: "Mountain Dew 12PK",
				Price:            "6.49",
			}, {
				ShortDescription: "Emils Cheese Pizza",
				Price:            "12.25",
			}, {
				ShortDescription: "Knorr Creamy Chicken",
				Price:            "1.26",
			}, {
				ShortDescription: "Doritos Nacho Cheese",
				Price:            "3.35",
			}, {
				ShortDescription: "Klarbrunn 12-PK 12 FL OZ",
				Price:            "12.00",
			},
		},
		Total: "35.35",
	}

	// post receipt
	body, err := json.Marshal(testReceipt)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Id string `json:"id"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Id)

	// get receipt
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/receipts/"+response.Id+"/points", nil)
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var fetchedReceipt receipt
	err = json.Unmarshal(recorder.Body.Bytes(), &fetchedReceipt)

	assert.NoError(t, err)
	assert.Equal(t, fetchedReceipt, testReceipt)

}
