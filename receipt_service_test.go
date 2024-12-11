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

var validTestCases = []struct {
	description    string
	receipt        receipt
	expectedPoints int
}{
	{
		description: "example1",
		receipt: receipt{
			Retailer:     "Target",
			PurchaseDate: "2022-01-01",
			PurchaseTime: "13:01",
			Items: []receiptItem{
				{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
				{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
				{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
				{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
				{ShortDescription: "Klarbrunn 12-PK 12 FL OZ", Price: "12.00"},
			},
			Total: "35.35",
		},
		expectedPoints: 28,
	},
	{
		description: "example2",
		receipt: receipt{
			Retailer:     "M&M Corner Market",
			PurchaseDate: "2022-03-20",
			PurchaseTime: "14:33",
			Items: []receiptItem{
				{ShortDescription: "Gatorade", Price: "2.25"},
				{ShortDescription: "Gatorade", Price: "2.25"},
				{ShortDescription: "Gatorade", Price: "2.25"},
				{ShortDescription: "Gatorade", Price: "2.25"},
			},
			Total: "9.00",
		},
		expectedPoints: 109,
	},
	{
		description: "time at exactly 2pm",
		receipt: receipt{
			Retailer:     "Walgreens",  // 9
			PurchaseDate: "2022-01-21", // 6
			PurchaseTime: "14:00",
			Items: []receiptItem{
				{ShortDescription: "Milk", Price: "4.00"},
				{ShortDescription: "Bread", Price: "6.00"},
			}, // 5
			Total: "10.00", // 50 + 25
		},
		expectedPoints: 95,
	},
	{
		description: "time at exactly 4pm and strange prices",
		receipt: receipt{
			Retailer:     "Super ABC",  // 8
			PurchaseDate: "2022-01-15", // 6
			PurchaseTime: "16:00",
			Items: []receiptItem{
				{ShortDescription: "  ABC  ", Price: "12.34"},            // 3
				{ShortDescription: "ABCDEF", Price: "0.00"},              // 0
				{ShortDescription: "ABCDEFGHI", Price: "93619473927.49"}, // 18723894786
			}, // 5
			Total: "93619473939.83",
		},
		expectedPoints: 18723894808,
	},
	{
		description: "0 points",
		receipt: receipt{
			Retailer:     "&&&-__- _ -& ",
			PurchaseDate: "2024-11-12",
			PurchaseTime: "00:00",
			Items: []receiptItem{
				{ShortDescription: "item", Price: "1.01"},
			},
			Total: "1.01",
		},
		expectedPoints: 0,
	},
}

var invalidTestCases = []struct {
	description string
	receipt     receipt
}{
	{
		description: "empty strings",
		receipt: receipt{
			Retailer:     "",
			PurchaseDate: "2024-11-12",
			PurchaseTime: "12:00",
			Items: []receiptItem{
				{ShortDescription: "", Price: "1.00"},
			},
			Total: "1.00",
		},
	},
	{
		description: "invalid date",
		receipt: receipt{
			Retailer:     "walmart",
			PurchaseDate: "2023-02-29",
			PurchaseTime: "12:00",
			Items: []receiptItem{
				{ShortDescription: "something", Price: "1.00"},
			},
			Total: "1.00",
		},
	},
	{
		description: "invalid time",
		receipt: receipt{
			Retailer:     "walmart",
			PurchaseDate: "2024-12-11",
			PurchaseTime: "25:99",
			Items: []receiptItem{
				{ShortDescription: "something", Price: "1.00"},
			},
			Total: "1.00",
		},
	},
	{
		description: "no items",
		receipt: receipt{
			Retailer:     "walmart",
			PurchaseDate: "2024-12-11",
			PurchaseTime: "12:00",
			Items:        []receiptItem{},
			Total:        "1.00",
		},
	},
	{
		description: "invalid dollar amount format",
		receipt: receipt{
			Retailer:     "walmart",
			PurchaseDate: "2024-12-11",
			PurchaseTime: "25:99",
			Items: []receiptItem{
				{ShortDescription: "something", Price: "50.0005"},
			},
			Total: "50.0005",
		},
	},
	{
		description: "missing total",
		receipt: receipt{
			Retailer:     "walmart",
			PurchaseDate: "2024-12-11",
			PurchaseTime: "25:99",
			Items: []receiptItem{
				{ShortDescription: "something", Price: "50.0005"},
			},
		},
	},
}

func setupTestContext() *gin.Engine {
	store := &receiptStore{
		receipts: make(map[string]receiptWrapper),
	}
	router := setupRouter(store)
	return router
}

func TestProcessAndReceive(t *testing.T) {
	router := setupTestContext()

	testReceipt := validTestCases[0].receipt
	expectedPoints := validTestCases[0].expectedPoints

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

	var pointsResponse struct {
		Points int `json:"points"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &pointsResponse)

	assert.NoError(t, err)
	assert.Equal(t, pointsResponse.Points, expectedPoints)

}

func TestAllValidReceipts(t *testing.T) {
	// post all receipts, then get all receipts in order
	router := setupTestContext()

	receiptIds := make([]string, len(validTestCases))

	for i, tc := range validTestCases {
		testReceipt := tc.receipt
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

		receiptIds[i] = response.Id
	}

	for i, tc := range validTestCases {
		id := receiptIds[i]
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/receipts/"+id+"/points", nil)
		router.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)

		var pointsResponse struct {
			Points int `json:"points"`
		}
		err := json.Unmarshal(recorder.Body.Bytes(), &pointsResponse)

		assert.NoError(t, err)
		assert.Equal(t, pointsResponse.Points, tc.expectedPoints)
	}

}

func TestInvalidReceipts(t *testing.T) {
	router := setupTestContext()

	for _, tc := range invalidTestCases {
		testReceipt := tc.receipt
		body, err := json.Marshal(testReceipt)
		assert.NoError(t, err)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, response.Error, "The receipt is invalid.")
	}
}

func TestNonexistentReceipt(t *testing.T) {
	router := setupTestContext()

	id := "this id is invalid"
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/receipts/"+id+"/points", nil)
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, response.Error, "No receipt found for that ID.")
}
