package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"

	_ "github.com/lib/pq"
)

// HalooDB is a local database client
type HalooDB struct {
	// The database connection.
	connection *sql.DB

	// The database insertion queue channel
	queue chan Message

	// Whether or not to run the initial migration
	runMigration bool
}

func newHalooDB(migrate bool) *HalooDB {
	return &HalooDB{
		queue:        make(chan Message),
		runMigration: migrate,
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
	if hdb.runMigration {
		hdb.migrate()
		err = hdb.test()

		if err == nil {
			fmt.Println("*** DATABASE TESTED AND WORKING ***")
		}
	}

	fmt.Println("*** HALOO CHAT RUNNING! :) ***")
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
			stmt, err := hdb.connection.Prepare("INSERT INTO chatlog (sender, receiver, message, room_id, timestamp) VALUES ($1, $2, $3, $4, $5)")

			if err != nil {
				log.Printf("error preparing message to db: %v", err)
			}

			_, err = stmt.Exec(message.Sender, message.Receiver, message.Message, message.RoomID, message.Timestamp)
			if err != nil {
				log.Printf("error inserting message to db: %v", err)
			}
		}
	}
}

func (hdb *HalooDB) start() {
	var err error
	if runtime.GOOS == "windows" {
		cmd := exec.Command("./bin/.\\cockroach", "start", "--insecure")
		err = cmd.Start()
	} else if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		cmd := exec.Command("./bin/cockroach", "start", "--insecure")
		err = cmd.Start()
	}

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

// Create default data for testing
func (hdb *HalooDB) createDefaultData() {
	var userID int
	var userTwoID int
	var roomID int

	if hdb.rowCount("chat_users") == 0 {
		// Insert default user into users table.
		if err := hdb.connection.QueryRow(
			"INSERT INTO chat_users (name, email, password, last_seen, profile_picture) VALUES ('Superadmin', 'admin@haloochat.dev', 'password', '2017-10-25 10:10:10.555555-05:00', 'admin.jpg') RETURNING id").Scan(&userID); err != nil {
			log.Printf("error inserting default user into users: %v", err)
		}

		if err := hdb.connection.QueryRow(
			"INSERT INTO chat_users (name, email, password, last_seen, profile_picture) VALUES ('Superadmin2', 'admin2@haloochat.dev', 'password2', '2017-10-25 10:10:10.555555-05:00', 'admin2.jpg') RETURNING id").Scan(&userTwoID); err != nil {
			log.Printf("error inserting default user into users: %v", err)
		}
	}

	if hdb.rowCount("rooms") == 0 {
		// Insert default room into rooms table.
		if err := hdb.connection.QueryRow("INSERT INTO rooms (name, picture) VALUES ('Welcome', 'placeholder.jpg') RETURNING id").Scan(&roomID); err != nil {
			log.Printf("error inserting default room to rooms: %v", err)
		}

		stmt, err := hdb.connection.Prepare("INSERT INTO room_has_users (room_id, user_id, is_admin) VALUES ($1, $2, $3)")

		if err != nil {
			log.Printf("error preparing foreign keys for default rooms: %v", err)
		}

		_, err = stmt.Exec(roomID, userID, true)
		if err != nil {
			log.Printf("error creating foreign keys for default rooms: %v", err)
		}

		stmt, err = hdb.connection.Prepare("INSERT INTO user_conversations (user_id, receiver_user_id) VALUES ($1, $2)")

		if err != nil {
			log.Printf("error preparing default data to user conversations: %v", err)
		}

		_, err = stmt.Exec(userID, userTwoID)
		if err != nil {
			log.Printf("error inserting default data to user conversations: %v", err)
		}
	}

	stmt, err := hdb.connection.Prepare("INSERT INTO chatlog (sender, receiver, message, room_id, timestamp) VALUES ($1, $2, $3, $4, $5)")

	if err != nil {
		log.Printf("error preparing chatlog data: %v", err)
	}

	_, err = stmt.Exec(userID, userTwoID, "Testataan kannan kautta tulevia viestejä", roomID, "1513012789379")
	if err != nil {
		log.Printf("error inserting chatlog data: %v", err)
	}

	stmt, err = hdb.connection.Prepare("INSERT INTO chatlog (sender, receiver, message, room_id, timestamp) VALUES ($1, $2, $3, null, $4)")
	if err != nil {
		log.Printf("error preparing chatlog data: %v", err)
	}

	_, err = stmt.Exec(userID, userTwoID, "Testataan kannan kautta tulevia priva viestejä", "1513012789379")
	if err != nil {
		log.Printf("error inserting chatlog data: %v", err)
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
