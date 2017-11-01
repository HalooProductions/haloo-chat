package main

import "log"

// Room is a hub of multiple chat users
type Room struct {
	ID      int
	Name    string
	Picture string
}

func getRooms(db *HalooDB) []Room {
	var rooms []Room

	rows, err := db.connection.Query("SELECT * FROM rooms;")
	if err != nil {
		log.Printf("error getting rooms from the db: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var room Room

		if err := rows.Scan(&room.ID, &room.Name, &room.Picture); err != nil {
			log.Printf("error reading room data from db: %v", err)
		}

		rooms = append(rooms, room)
	}

	return rooms
}
