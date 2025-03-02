package main

import (
	"database/sql"
	"log"
	"math/rand"
	"os"

	"github.com/harshjoeyit/flickr-ticket/db"
)

type Ticket struct {
	DBs []*sql.DB
}

func NewTicket() *Ticket {
	t := &Ticket{}

	t.DBs = make([]*sql.DB, 2)

	var err error

	t.DBs[0], err = db.Connect("root", "mysql", "localhost", 3306, "test")
	if err != nil {
		panic(err)
	}

	t.DBs[1], err = db.Connect("root", "mysql", "localhost", 3307, "test")
	if err != nil {
		panic(err)
	}

	return t
}

func (t *Ticket) loadBalancer() *sql.DB {
	return t.DBs[rand.Intn(2)]
}

func (t *Ticket) NewID() (id int64, err error) {
	db := t.loadBalancer()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return 0, err
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit / Rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
			log.Println("Error during transaction, rolling back:", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Println("Error committing transaction:", err)
			}
		}
	}()

	// incrementing by 2, since we have two server, one for odd IDs and one for even IDs
	query := "INSERT INTO ticket (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = id + 2;"
	// query := "INSERT INTO ticket (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id + 1);"

	// res, err := db.Exec(query)
	res, err := tx.Exec(query)
	if err != nil {
		return
	}

	id, err = res.LastInsertId()
	if err != nil {
		return
	}

	return id, nil
}

func setupLogger() {
	logFile, err := os.OpenFile("flickr.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Include timestamp and line number
}

func main() {
	// setupLogger()

	t := NewTicket()

	for range 1000 {
		id, err := t.NewID()
		if err != nil {
			log.Println(err)
			continue
		}

		_ = id
		// log.Println(id)
	}
}
