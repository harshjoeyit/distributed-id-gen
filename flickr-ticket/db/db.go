package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	maxOpenConn = 10
)

// ConnectDB establishes a connection to the database
func Connect(user, pswd, host string, port int, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pswd, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Set database connection pooling options
	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping to ensure the connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return nil, err
	}

	// fmt.Printf("Connected to '%s'.\n", dbName)
	return db, nil
}
