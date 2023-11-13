package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Method string

const (
	GET  Method = "get"
	POST Method = "post"
)

// LoadTestingRequestBody :- Load test API paylaod structure
type LoadTestingRequestBody struct {
	URL         string                  `json:"url" binding:"required"`
	Method      Method                  `json:"method" binding:"required,eq=get|eq=post"`
	Parallelism int                     `json:"parallelism" binding:"required"`
	LoadVolume  int                     `json:"loadVolume" binding:"required"`
	Headers     map[string][]string     `json:"headers" binding:"required"`
	Body        *map[string]interface{} `json:"body"`
}

// simulateLoadTest :- Handler for Load Test API.
func (payload *LoadTestingRequestBody) simulateLoadTest() {

	// Create a channel of type *LoadTestingRequestBody
	loadTestChannel := make(chan *LoadTestingRequestBody)

	// Opening Go threads to process payload and make HTTP calls as per the `Parallelism` passed
	for i := 1; i <= payload.Parallelism; i++ {
		go loadChannelListner("GoRoutine_"+strconv.Itoa(i), loadTestChannel)
	}

	// Send values to the channel as per the `LoadVolume` passed
	for i := 1; i <= payload.LoadVolume; i++ {
		loadTestChannel <- payload
	}

	// Close the channel to signal that no more values will be sent
	close(loadTestChannel)

	// Wait for a moment to allow goroutines to finish processing
	time.Sleep(time.Second)

}

func loadChannelListner(goroutineName string, ch chan *LoadTestingRequestBody) {
	for {
		// Receive a LoadTest payload from the channel
		value, ok := <-ch

		// Check if the channel is closed
		if !ok {
			fmt.Printf("%s: Channel closed\n", goroutineName)
			return
		}

		// Check whether load testing API a GET call or POST call, accordingly call http method wrappers.
		if value.Method == GET {
			// Make a GET call
			makeGetCall(value.URL, value.Headers)
		} else {
			// Make a POST call
			makePostCall(value.URL, value.Body, value.Headers)
		}
	}
}

func makeGetCall(url string, headers map[string][]string) {

	// Make a GET request to the external API
	// Create a new GET request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set custom headers
	request.Header = headers

	// Use the default HTTP client to send the request
	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return
	}

	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading the response:", err)
		return
	}

	// Print the response body
	fmt.Println("Response from the API:")
	fmt.Println(string(body))

}

func makePostCall(url string, payload *map[string]interface{}, headders map[string][]string) {

	// Convert the map to a JSON-formatted byte array
	byteArray, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create a new POST request with the JSON payload
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(byteArray))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	request.Header = headders

	// Use the default HTTP client to send the request
	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return
	}

	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading the response:", err)
		return
	}

	// Print the response body
	fmt.Println("Response from the API:")
	fmt.Println(string(body))
}

func main() {
	// Create a Gin router with default middleware
	router := gin.Default()

	// Define routes

	// Define /hello GET route
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, Gin!"})
	})

	// Define /simulate/load_testing POST route with the request body structure
	router.POST("/simulate/load_testing", func(c *gin.Context) {
		var requestBody LoadTestingRequestBody

		// Parse JSON from the request body into the defined struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parallelism should be a positive integer with value less than equal to 1000.
		if requestBody.Parallelism <= 0 || requestBody.Parallelism > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parallelism should be between 0, 1000"})
			return
		}

		// call simulateLoadTest in non-blocking mannger.
		go requestBody.simulateLoadTest()

		// Process the JSON data
		// In this example, we simply echo the received data
		c.JSON(http.StatusOK, gin.H{"data": requestBody})
	})

	// Run the server on port 8080
	router.Run(":8080")
}
