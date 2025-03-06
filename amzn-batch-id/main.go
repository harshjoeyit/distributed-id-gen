package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"github.com/harshjoeyit/amzn-batch-id/db"
)

var DB *sql.DB

const HTTPServerPort = 8080
const IDRange = 5 // 100, 500, 1000, depending on the use case

type IDBatch struct {
	st int64
	en int64
}

func getNewIDBatch(serviceName string) (idBatch *IDBatch, err error) {
	log.Printf("Generating new ID batch for service: %s", serviceName)

	tx, err := DB.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return
	}

	// Acquire exclusive lock
	var counter int64
	query := "SELECT counter FROM amazon_id WHERE service_name = ? FOR UPDATE;"

	err = tx.QueryRow(query, serviceName).Scan(&counter)
	if err != nil {
		tx.Rollback()
		return
	}

	// Update counter
	query = "UPDATE amazon_id SET counter = counter + ? WHERE service_name = ?;"

	_, err = tx.Exec(query, IDRange, serviceName)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return
	}

	return &IDBatch{
		st: counter + 1,
		en: counter + IDRange,
	}, nil
}

func isValidServiceName(serviceName string) bool {
	// validation logic
	return serviceName == "order" || serviceName == "product"
}

func main() {
	// Connect to database
	var err error
	DB, err = db.Connect("root", "mysql", "localhost", 3306, "test")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	go client("order", "s1")
	go client("order", "s2")

	ch := make(chan struct{})

	<-ch

	// Expose REST API

	/*
		ge := gin.Default()

		ge.POST("api/id", func(c *gin.Context) {
			type reqBody struct {
				ServiceName string `json:"service_name" binding:"required"`
			}

			var b reqBody
			if err := c.ShouldBindJSON(&b); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if !isValidServiceName(b.ServiceName) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service name"})
				return
			}

			idBatch, err := getNewIDBatch(b.ServiceName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(
				http.StatusOK,
				gin.H{"start_id": idBatch.st, "end_id": idBatch.en},
			)
		})

		ge.Run(fmt.Sprintf(":%d", HTTPServerPort))
	*/
}

func client(serviceName string, instanceID string) {
	// Get the first batch of IDs
	idBatch, err := getNewIDBatch(serviceName)
	if err != nil {
		log.Println(err)
		return
	}

	currID := idBatch.st

	for {
		// Do some processing
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

		if currID > idBatch.en {
			// Batch is exhausted, get a new batch
			log.Printf("%s - %s exhausted batch [%d, %d]\n", serviceName, instanceID, idBatch.st, idBatch.en)

			idBatch, err = getNewIDBatch(serviceName)
			if err != nil {
				log.Println(err)
				return
			}

			currID = idBatch.st
		}

		log.Printf("%s - %s used id: %d\n", serviceName, instanceID, currID)

		currID++
	}
}
