package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshjoeyit/my-snowflake/machineid"
)

const HTTPServerPort int = 5001

// Snowflake ID
// [1 bit][41 bits   ][10 bits   ][12 bits        ]
// [empty][epoch (ms)][machine ID][sequence number]

type Snowflake struct {
	MachineID     int
	LastTimestamp int64
	Sequence      int
	Mu            sync.Mutex // Mutex to protect shared variables (lastTimestamp, sequence)
}

func NewSnowflake() *Snowflake {
	s := &Snowflake{}

	machineID, err := machineid.Get()
	if err != nil {
		panic(err)
	}
	// log.Printf("Machine ID: %d", machineID) // Log machine ID

	s.MachineID = machineID
	s.Mu = sync.Mutex{}

	return s
}

func (s *Snowflake) GenerateNewID() (int64, error) {
	s.Mu.Lock()         // Acquire the lock
	defer s.Mu.Unlock() // Release the lock when the function returns

	timestamp := time.Now().UnixMilli()

	if timestamp < s.LastTimestamp {
		return 0, fmt.Errorf("clock has drifted backwards, rejecting request until %d", s.LastTimestamp)
	}

	// Get sequence number
	if s.LastTimestamp == timestamp {
		// increment sequence
		s.Sequence = (s.Sequence + 1) & 4095 // sequence mask

		if s.Sequence == 0 {
			// happens when 4096 & 4095, which is when the sequence number
			// of current timestamp have exhauseted, we need timestamp
			// of next millisecond
			timestamp = s.nextMillis(s.LastTimestamp)
		}
	} else {
		// reset last timestamp and sequence
		s.LastTimestamp = timestamp
		s.Sequence = 0
	}

	// // Unix epoch time in milliseconds. Binary repesentation takes 41 bits
	// bin := fmt.Sprintf("%b", timestamp)
	// log.Printf("epoch now(ms): %d, binary: %s, bits: %d\n", timestamp, bin, len(bin))

	// // machine ID is 10 bits
	// bin = fmt.Sprintf("%b", s.MachineID)
	// log.Printf("machine ID: %d, binary: %s, bits: %d\n", s.MachineID, bin, len(bin))

	// bin = fmt.Sprintf("%b", s.Sequence)
	// log.Printf("sequence Number: %d, binary: %s, bits: %d\n", s.Sequence, bin, len(bin))

	id := (timestamp << 22) | (int64(s.MachineID) << 12) | int64(s.Sequence)

	// bin = fmt.Sprintf("%b", id)
	// log.Printf("id: %d, bin %s, bits: %d\n\n", id, bin, len(bin))

	return id, nil
}

// nextMillis does busy waiting for the clock to reach next millisecond
// and returns it
func (s *Snowflake) nextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixMilli()

	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixMilli()
	}

	return timestamp
}

func setupLogger() {
	logFile, err := os.OpenFile("snowflake.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Include timestamp and line number
}

func main() {
	setupLogger() // Initialize logger

	s := NewSnowflake()

	ge := gin.Default()

	ge.GET("api/snowflake", func(c *gin.Context) {
		id, err := s.GenerateNewID()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"id": id})
	})

	ge.Run(fmt.Sprintf(":%d", HTTPServerPort))
}
