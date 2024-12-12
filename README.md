# Fetch Receipt Processor Challenge

### Building and running from source
1. Initialize the Go module and install dependencies:   ```go mod tidy```
2. Build and run the program (choose one method):
   
   **Method 1: Direct Run**   ```go run .   ```

   **Method 2: Build and Execute**   
   ```go build```\
   ```./receipt-processor  # On Unix/Linux/Mac```\
   **or**\
   ```.\receipt-processor.exe  # On Windows   ```

To run the service on a port different from 8080 (the default), run with the flag `--port [port]`


### Running Test Cases
Test cases have been provided in ```receipt_service_test.go```. Run them with ```go test```.

### Interacting with the web service

```bash
curl http://localhost:8080/receipts/process \
  --include \
  --header "Content-Type: application/json" \
  --request "POST" \
  --data '{"retailer":"Target","purchaseDate":"2022-01-01","purchaseTime":"13:01","items":[{"shortDescription":"Mountain Dew 12PK","price":"6.49"},{"shortDescription":"Emils Cheese Pizza","price":"12.25"},{"shortDescription":"Knorr Creamy Chicken","price":"1.26"},{"shortDescription":"Doritos Nacho Cheese","price":"3.35"},{"shortDescription":"   Klarbrunn 12-PK 12 FL OZ  ","price":"12.00"}],"total":"35.35"}'
```

```bash
curl http://localhost:8080/receipts/270e90bb-2c3f-4e70-ba0a-41345e784db0/points
```


## Explanation of important design decisions

### Computation of points immediately after receipt is processed
In the present design, the amount of points assigned to a receipt is computed and saved immediately after a POST request is received. This is in contrast to computing the amount of points "on-demand" only when an ID query is received. This decision was made because the receipts are immutable in this API specification, and it is assumed that GET requests will be more frequent than POST requests. If in a real system these assumptions are violated, an alternative design where the points are computed and possibly cached on ID query could be considered.

### In-memory data store
A map was chosen as the data store for the receipts for its simplicity and constant time lookup. It was wrapped in a `receiptStore` struct to allow for dependency injection for separation between test and production runtime environments.

### Single package organization
All code is contained within the `main` package. This is for simplicity for the purposes of this exercise. A more complicated production implementation would separate functional components into different packages.