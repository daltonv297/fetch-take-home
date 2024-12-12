package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.Int("port", 8080, "Port to run server on")
	flag.Parse()

	store := &receiptStore{
		receipts: make(map[string]receiptWrapper),
	}
	var router *gin.Engine = setupRouter(store)

	router.Run(fmt.Sprintf("localhost:%d", *port))
}
