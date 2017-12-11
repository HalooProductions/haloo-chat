package main

import (
	"log"
)

// User represents a single chat user
type User struct {
	ID             int    `json:"ID"`
	Name           string `json:"Name"`
	Email          string `json:"Email"`
	Password       string
	LastSeen       string `json:"Last_seen"`
	ProfilePicture string `json:"Picture"`
	DB             *HalooDB
}

// Find user from the database
func getUser(db *HalooDB, id int) User {
	var user User

	rows, err := db.connection.Query("SELECT id, name, email, password, last_seen, profile_picture FROM chat_users WHERE id = $1;", id)

	if err != nil {
		log.Printf("error getting user from db: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.LastSeen, &user.ProfilePicture); err != nil {
			log.Printf("error reading user data from database: %v", err)
		}
	}

	user.DB = db

	return user
}

// Get all conversations for one user
func (user *User) getConversations() []User {
	var conversations []User

	rows, err := user.DB.connection.Query("SELECT id, name, email, last_seen, profile_picture FROM chat_users c WHERE c.id IN (SELECT receiver_user_id FROM user_conversations WHERE user_id = $1);", user.ID)

	if err != nil {
		log.Printf("error getting user conversations from db: %v", err)
	}

	log.Printf("rows: %v", rows)

	defer rows.Close()
	for rows.Next() {
		var conversation User

		if err := rows.Scan(&conversation.ID, &conversation.Name, &conversation.Email, &conversation.LastSeen, &conversation.ProfilePicture); err != nil {
			log.Printf("error reading conversation from db: %v", err)
		}

		conversations = append(conversations, conversation)
	}

	return conversations
}

// Get all rooms for one user
func (user *User) getRooms() []Room {
	var rooms []Room

	rows, err := user.DB.connection.Query("SELECT id, name, picture FROM rooms r WHERE id IN (SELECT room_id FROM room_has_users rh WHERE rh.user_id = $1);", user.ID)
	if err != nil {
		log.Printf("error getting user rooms from db: %v", err)
	}

	log.Printf("room rows: %v", rows)

	defer rows.Close()
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Picture); err != nil {
			log.Printf("error reading room for user rooms from db: %v", err)
		}

		rooms = append(rooms, room)
	}

	return rooms
}

/*func (user *User) joinRoom() bool {
	stmt, err := user.DB.connection.Prepare("INSERT INTO room_has_users (room_id, user_id) VALUES ($1, $2)")
}*/
