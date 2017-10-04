package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	_ "github.com/lib/pq"
)

// HalooDB is a local database client
type HalooDB struct {
	// The database connection.
	connection *sql.DB

	// The database insertion queue channel
	queue chan Message
}

func newHalooDB() *HalooDB {
	return &HalooDB{
		queue: make(chan Message),
	}
}

func (hdb *HalooDB) connect() {
	// Connect to the "haloochat" database.
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/haloochat?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	hdb.connection = db
	go hdb.start()
	hdb.migrate()
	err = hdb.test()

	if err != nil {
		log.Printf("error testing the database: %v", err)
		return
	}

	fmt.Println("*** DATABASE TESTED AND WORKING ***")
}

func (hdb *HalooDB) test() error {
	var err error

	// Insert test user into the users table.
	if _, err = hdb.connection.Exec(
		"INSERT INTO users (name, email, password, last_seen, profile_picture) VALUES ('Testuser', 'test@gmail.com', 'password', '2016-01-25 10:10:10.555555-05:00', 'test.jpg')"); err != nil {
		log.Printf("error inserting to users: %v", err)
	}

	if hdb.rowCount("users") == 0 {
		err = errors.New("No test user found in the database")
	}

	rows, err := hdb.connection.Query("SELECT name, email, password FROM users")
	if err != nil {
		log.Printf("error querying test data from users: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var userName, email, password string
		if err = rows.Scan(&userName, &email, &password); err != nil {
			log.Fatal(err)
		}
	}

	if _, err = hdb.connection.Exec(
		"DELETE FROM users WHERE name LIKE 'Testuser'"); err != nil {
		log.Printf("error deleting from users: %v", err)
	}

	return err
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

func (hdb *HalooDB) queuePump() {
	for {
		select {
		case message := <-hdb.queue:
			stmt, err := hdb.connection.Prepare("INSERT INTO chatlog (sender, message, timestamp) VALUES ($1, $2, $3)")

			if err != nil {
				log.Printf("error preparing message to db: %v", err)
			}

			timestamp := time.Unix(message.Timestamp, 0)

			_, err = stmt.Exec(message.Sender, message.Message, timestamp.Format(time.RFC3339))
			if err != nil {
				log.Printf("error inserting message to db: %v", err)
			}
		}
	}
}

func (hdb *HalooDB) start() {
	cmd := exec.Command("./bin/.\\cockroach", "start", "--insecure", "--host=localhost")
	log.Printf("Running command and waiting for it to finish...")
	err := cmd.Start()
	log.Printf("Command finished with error: %v", err)
}

func (hdb *HalooDB) migrate() {
	data, err := ioutil.ReadFile("./database/migration.sql")
	if err != nil {
		log.Printf("error reading migration file: %v", err)
	}

	dataStr := string(data)

	_, err = hdb.connection.Exec(dataStr)
	if err != nil {
		log.Printf("error executing the migration: %v", err)
	}
}
