package main

import "log"

// User represents a single chat user
type User struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Password       string
	LastSeen       string `json:"last_seen"`
	ProfilePicture string `json:"profile_picture"`
}

func getUser(db *HalooDB, id int) User {
	var user User

	rows, err := db.connection.Query("SELECT id, name, email, password, last_seen, profile_picture FROM chat_users WHERE id = ?;", id)

	if err != nil {
		log.Printf("error getting user from db: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.LastSeen, &user.ProfilePicture); err != nil {
			log.Printf("error reading user data from database: %v", err)
		}
	}

	return user
}

func (user *User) getConversations(db *HalooDB) []User {
	var conversations []User

	rows, err := db.connection.Query("SELECT id, name, email, last_seen, profile_picture FROM chat_users c WHERE c.id IN (SELECT receiver_user_id FROM user_conversations WHERE user_id = ?);", user.ID)

	if err != nil {
		log.Printf("error getting user conversations from db: %v", err)
	}

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

func (user *User) getRooms(db *HalooDB) []Room {
	var rooms []Room

	rows, err := db.connection.Query("SELECT * from rooms r WHERE id IN (SELECT room_id FROM room_has_users rh WHERE rh.user_id = ?);", user.ID)
	if err != nil {
		log.Printf("error getting user rooms from db: %v", err)
	}

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
