package main

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type receiptItem struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type receipt struct {
	Retailer     string        `json:"retailer"`
	PurchaseDate string        `json:"purchaseDate"`
	PurchaseTime string        `json:"purchaseTime"`
	Items        []receiptItem `json:"items"`
	Total        string        `json:"total"`
}

type receiptWrapper struct {
	Receipt receipt `json:"receipt"`
	Points  int     `json:"points"`
}

type receiptStore struct {
	receipts map[string]receiptWrapper
}

func (store *receiptStore) getReceiptPointsById(c *gin.Context) {
	// GET: /receipts/{id}/points
	id := c.Param("id")

	receiptGet, ok := store.receipts[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No receipt found for that ID.",
		})
		return
	}

	returnPointsJSON := struct {
		Points int `json:"points"`
	}{Points: receiptGet.Points}

	c.IndentedJSON(http.StatusOK, returnPointsJSON)
}

func (store *receiptStore) processReceipts(c *gin.Context) {
	// POST: /receipts/process

	var newReceipt receipt
	if err := c.BindJSON(&newReceipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The receipt is invalid.",
		})
		return
	}

	// validate receipt field formats
	valid := regexp.MustCompile(`^[\w\s\-&]+$`).MatchString(newReceipt.Retailer) &&
		regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(newReceipt.Total) &&
		len(newReceipt.Items) > 0

	for _, item := range newReceipt.Items {
		valid = valid && regexp.MustCompile(`^[\w\s\-]+$`).MatchString(item.ShortDescription) &&
			regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(item.Price)
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The receipt is invalid.",
		})
		return
	}

	// assumes duplicate receipts with identical content are allowed
	id := uuid.New().String()
	receiptPoints, err := computePoints(&newReceipt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The receipt is invalid.",
		})
		return
	}
	store.receipts[id] = receiptWrapper{Receipt: newReceipt, Points: receiptPoints}

	returnIdJSON := struct {
		Id string `json:"id"`
	}{Id: id}

	c.IndentedJSON(http.StatusOK, returnIdJSON)
}

func computePoints(receipt *receipt) (int, error) {
	points := 0
	points += countAlphanumeric(receipt.Retailer)

	totalLastThree := receipt.Total[len(receipt.Total)-3:]

	// string match to avoid floating point error
	if totalLastThree == ".00" {
		points += 50
	}
	if arr := []string{".00", ".25", ".50", ".75"}; slices.Contains(arr, totalLastThree) {
		points += 25
	}

	points += 5 * (len(receipt.Items) / 2)

	for _, item := range receipt.Items {
		if utf8.RuneCountInString(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			priceFloat, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid price format: %v", err)
			}
			points += int(math.Ceil(priceFloat * 0.2))
		}
	}

	dateLayout := "2006-01-02"
	timeLayout := "15:04"
	purchaseDate, err := time.Parse(dateLayout, receipt.PurchaseDate)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %v", err)
	}
	purchaseTime, err := time.Parse(timeLayout, receipt.PurchaseTime)
	if err != nil {
		return 0, fmt.Errorf("invalid time format: %v", err)
	}

	if purchaseDate.Day()%2 == 1 {
		points += 6
	}

	timeMin, _ := time.Parse(timeLayout, "14:00")
	timeMax, _ := time.Parse(timeLayout, "16:00")

	if purchaseTime.After(timeMin) && purchaseTime.Before(timeMax) {
		points += 10
	}

	return points, nil
}

func countAlphanumeric(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

func setupRouter(store *receiptStore) *gin.Engine {
	router := gin.Default()
	router.GET("/receipts/:id/points", store.getReceiptPointsById)
	router.POST("/receipts/process", store.processReceipts)
	return router
}
