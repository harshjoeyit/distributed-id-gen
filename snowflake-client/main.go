package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

const numGoRoutines = 100
const numReqPerGoRoutine = 100

type SnowflakeResponse struct {
	Id int64 `json:"id"`
}

func request(f *os.File, routineID, reqNum int) {
	resp, err := http.Get("http://localhost:5001/api/snowflake")
	if err != nil {
		log.Printf("Error making request (routine %d, request %d): %v", routineID, reqNum, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body (routine %d, request %d): %v", routineID, reqNum, err)
		return
	}

	// Extract ID from the JSON response
	var snowflakeResponse SnowflakeResponse
	err = json.Unmarshal(body, &snowflakeResponse)
	if err != nil {
		log.Printf("Error unmarshaling JSON (routine %d, request %d): %v", routineID, reqNum, err)
		return
	}

	// Write the ID to a file
	if _, err := f.Write([]byte(fmt.Sprintf("%d\n", snowflakeResponse.Id))); err != nil {
		log.Printf("Error writing to file (routine %d, request %d): %v", routineID, reqNum, err)
		return
	}
	// fmt.Printf("Routine %d: Appended snowflake ID %d to snowflake_id.txt\n", routineID, reqNum)
}

func makeRequests(routineID int) {
	// Append to the file
	f, err := os.OpenFile("snowflake_id.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file (routine %d): %v", routineID, err)
		return
	}
	defer f.Close()

	for j := range numReqPerGoRoutine {
		request(f, routineID, j+1)
	}
}

func main() {
	var wg sync.WaitGroup

	for i := range numGoRoutines {
		wg.Add(1) // Increment the WaitGroup counter

		go func(routineID int) {
			makeRequests(i)

			wg.Done() // Decrement the WaitGroup counter when the goroutine completes
		}(i) // Pass the goroutine ID to the anonymous function
	}

	wg.Wait() // Wait for all goroutines to complete
	fmt.Println("All goroutines completed.")
}
