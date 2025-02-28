package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Snowflake ID
// [1 bit][41 bits   ][10 bits   ][12 bits        ]
// [empty][epoch (ms)][machine ID][sequence number]

var (
	machineID     int
	lastTimestamp int64
	sequence      int
	mu            sync.Mutex // Mutex to protect shared variables
)

const HTTPServerPort int = 5001

// getMachineID returns the machine ID (10 bits) of the current machine
func getMachineID() (int, error) {
	// return getMachineIDFromCentralService()
	// return 1, nil
	return hashIPtoMachineID()
}

// Hashes an IP address to generate a machine ID (10 bits)
func hashIPtoMachineID() (int, error) {
	// find private IP address of the machine
	ip, err := getPrivateIP()
	if err != nil {
		return 0, err
	}

	// fast, non-cryptographic hash
	h := fnv.New32a()
	h.Write([]byte(ip))
	mID := int(h.Sum32() % 1024) // 10 bit ID

	return mID, nil
}

// isPrivateIP checks if an IP address is private
func isPrivateIP(ip net.IP) bool {
	privateIPv4Ranges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateIPv4Ranges {
		_, ipNet, _ := net.ParseCIDR(cidr)

		if ipNet.Contains(ip) {
			return true
		}
	}

	privateIPv6Range := "fc00::/7"
	_, ipNet, _ := net.ParseCIDR(privateIPv6Range)
	return ipNet.Contains(ip)
}

func getPrivateIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("failed to get network interfaces: %v", err)
	}

	var privateIPs []string

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			if isPrivateIP(ip) {
				privateIPs = append(privateIPs, ip.String())
			}
		}
	}

	if len(privateIPs) == 0 {
		return "", nil
	}

	return privateIPs[0], nil
}

func generateNewID() (int64, error) {
	mu.Lock()         // Acquire the lock
	defer mu.Unlock() // Release the lock when the function returns

	timestamp := time.Now().UnixMilli()

	if timestamp < lastTimestamp {
		return 0, fmt.Errorf("clock has drifted backwards, rejecting request until %d", lastTimestamp)
	}

	// Get sequence number
	if lastTimestamp == timestamp {
		// increment sequence
		sequence = (sequence + 1) & 4095 // sequence mask

		if sequence == 0 {
			// happens when 4096 & 4095, which is when the sequence number
			// of current timestamp have exhauseted, we need timestamp
			// of next millisecond
			timestamp = nextMillis(lastTimestamp)
		}
	} else {
		// reset last timestamp and sequence
		lastTimestamp = timestamp
		sequence = 0
	}

	// Unix epoch time in milliseconds. Binary repesentation takes 41 bits
	bin := fmt.Sprintf("%b", timestamp)
	log.Printf("epoch now(ms): %d, binary: %s, bits: %d\n", timestamp, bin, len(bin))

	// machine ID is 10 bits
	bin = fmt.Sprintf("%b", machineID)
	log.Printf("machine ID: %d, binary: %s, bits: %d\n", machineID, bin, len(bin))

	bin = fmt.Sprintf("%b", sequence)
	log.Printf("sequence Number: %d, binary: %s, bits: %d\n", sequence, bin, len(bin))

	id := (timestamp << 22) | (int64(machineID) << 12) | int64(sequence)

	bin = fmt.Sprintf("%b", id)
	log.Printf("id: %d, bin %s, bits: %d\n\n", id, bin, len(bin))

	return id, nil
}

// nextMillis does busy waiting for the clock to reach next millisecond
// and returns it
func nextMillis(lastTimestamp int64) int64 {
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

	var err error

	machineID, err = getMachineID()
	if err != nil {
		panic(err)
	}
	log.Printf("Machine ID: %d", machineID) // Log machine ID

	ge := gin.Default()

	ge.GET("api/snowflake", func(c *gin.Context) {
		id, err := generateNewID()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"id": id})
	})

	ge.Run(fmt.Sprintf(":%d", HTTPServerPort))
}
