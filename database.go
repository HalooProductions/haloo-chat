package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// HalooDB is a local database client
type HalooDB struct {
	connection *sql.DB
}

func newHalooDB() *HalooDB {
	return &HalooDB{}
}

func (hdb *HalooDB) connect() {
	// Connect to the "haloochat" database.
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/haloochat?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	hdb.connection = db

	// Create the "users" table.
	if _, err := hdb.connection.Exec(
		"CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name VARCHAR(255), email VARCHAR(255), password VARCHAR(255), last_seen TIMESTAMPTZ, profile_picture VARCHAR(255))"); err != nil {
		log.Fatal(err)
	}

	if hdb.rowCount("users") == 0 {
		// Insert test user into the users table.
		if _, err := hdb.connection.Exec(
			"INSERT INTO users (name, email, password, last_seen, profile_picture) VALUES ('Riku', 'rikuw', 'salasana', '2016-01-25 10:10:10.555555-05:00', 'test.jpg')"); err != nil {
			log.Fatal(err)
		}
	}
}

func (hdb *HalooDB) rowCount(table string) int {
	var count int

	row := hdb.connection.QueryRow("SELECT COUNT(*) FROM " + table)
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	return count
}
