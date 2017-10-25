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
	hdb.test()
}

// Test the database
func (hdb *HalooDB) test() error {
	var err error

	// Insert test user into the users table.
	if _, err = hdb.connection.Exec(
		"INSERT INTO chat_users (name, email, password, last_seen, profile_picture) VALUES ('Testuser', 'test@gmail.com', 'password', '2016-01-25 10:10:10.555555-05:00', 'test.jpg')"); err != nil {
		log.Printf("error inserting to users: %v", err)
	}

	if hdb.rowCount("chat_users") == 1 {
		err = errors.New("No test user found in the database")
	}

	rows, err := hdb.connection.Query("SELECT name, email, password FROM chat_users")
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
		"DELETE FROM chat_users WHERE name LIKE 'Testuser'"); err != nil {
		log.Printf("error deleting from users: %v", err)
	}

	fmt.Println("*** DATABASE TESTED AND WORKING ***")

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
	err := cmd.Start()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}
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

	defer hdb.createDefaultData()
}

func (hdb *HalooDB) createDefaultData() {
	if hdb.rowCount("chat_users") == 0 {
		// Insert default user into users table.
		if _, err := hdb.connection.Exec(
			"INSERT INTO chat_users (name, email, password, last_seen, profile_picture) VALUES ('Superadmin', 'admin@haloochat.dev', 'password', '2017-10-25 10:10:10.555555-05:00', 'admin.jpg')"); err != nil {
			log.Printf("error inserting default user into users: %v", err)
		}
	}

	if hdb.rowCount("rooms") == 0 {
		// Insert default room into rooms table.
		if _, err := hdb.connection.Exec(
			"INSERT INTO rooms (name) VALUES ('Welcome')"); err != nil {
			log.Printf("error inserting to rooms: %v", err)
		}
	}
}

func (hdb *HalooDB) force() {
	data, err := ioutil.ReadFile("./database/force.sql")
	if err != nil {
		log.Printf("error reading migration file: %v", err)
	}

	dataStr := string(data)

	_, err = hdb.connection.Exec(dataStr)
	if err != nil {
		log.Printf("error executing the migration: %v", err)
	}
}
