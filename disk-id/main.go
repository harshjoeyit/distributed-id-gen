package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type Counter struct {
	Val      int64 // counter
	Mu       *sync.Mutex
	Filename string
}

func NewCounter(filename string) *Counter {
	val, err := Load(filename)
	if err != nil {
		panic(err)
	}

	return &Counter{
		Val:      val,
		Mu:       &sync.Mutex{},
		Filename: filename,
	}
}

func GetMachineID() int64 {
	return 42
}

// Load reads the counter value from the file on disk
func Load(filename string) (val int64, err error) {
	// Read counter value from file
	data, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("error reading file: %w", err)
		return
	}

	val, err = strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		err = fmt.Errorf("error parsing counter value: %w", err)
		return
	}

	return
}

// IncrAndSave increments counter value by 1 and flushes the counter value
// to file on disk
func (c *Counter) Save() error {
	// Convert the counter value to a string
	data := strconv.FormatInt(int64(c.Val), 10)

	err := os.WriteFile(c.Filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	// log.Printf("File saved with counter: %s\n", data)

	return nil
}

func (c *Counter) GetNewID() (id string, err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	c.Val++ // Increment to generate new ID

	// flush every nth value disk
	if c.Val%100 == 0 {
		err = c.Save()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	id = fmt.Sprintf("%d-%d", GetMachineID(), c.Val)

	return
}

func main() {
	c := NewCounter("counter.txt")

	var err error
	var id string

	// Test ID generation with go routines

	var wg sync.WaitGroup

	numGoRoutines := 10
	idsPerGoRoutine := 10

	wg.Add(numGoRoutines)

	for i := range numGoRoutines {
		go func(name string) {
			for range idsPerGoRoutine {
				id, err = c.GetNewID()
				if err != nil {
					log.Println(err)
					return
				}

				// _ = id
				fmt.Printf("%s: %s \n", name, id)
			}

			wg.Done()
		}(fmt.Sprintf("go%d", i))
	}

	wg.Wait()
}
