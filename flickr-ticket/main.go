package main

import (
	"database/sql"
	"log"
	"math/rand"
	"sync"

	"github.com/harshjoeyit/flickr-ticket/db"
)

const numDBServers = 2

type Ticket struct {
	DBs []*sql.DB
}

func NewTicket() *Ticket {
	t := &Ticket{}

	t.DBs = make([]*sql.DB, numDBServers)

	var err error

	// generates even IDs - 4,6,8,...
	t.DBs[0], err = db.Connect("root", "mysql", "localhost", 3306, "test")
	if err != nil {
		panic(err)
	}

	// generates odd IDs - 3,5,7,...
	t.DBs[1], err = db.Connect("root", "mysql", "localhost", 3307, "test")
	if err != nil {
		panic(err)
	}

	return t
}

func (t *Ticket) loadBalancer() *sql.DB {
	return t.DBs[rand.Intn(numDBServers)]
}

func (t *Ticket) NewID() (id int64, err error) {
	db := t.loadBalancer()

	// incrementing by 2, since we have two server, one for odd IDs and one for even IDs
	query := "INSERT INTO ticket (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = id + 2;"
	// query := "REPLACE INTO ticket (stub) VALUES ('a');"

	res, err := db.Exec(query)
	if err != nil {
		return
	}

	id, err = res.LastInsertId()
	if err != nil {
		return
	}

	return id, nil
}

func main() {
	t := NewTicket()

	numGoroutines := 10

	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(idx int) {
			defer wg.Done()

			id, err := t.NewID()

			if err != nil {
				log.Printf("error in goroutine: %d, %v\n", idx, err)
				return
			}

			log.Printf("goroutine: %d, id: %d\n", idx, id)

		}(i)
	}

	wg.Wait()
	log.Println("Done")
}
