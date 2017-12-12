// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
)

var addr = flag.String("addr", ":8000", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	// If path other than /chat, 404 error
	if r.URL.Path != "/chat" {
		http.Error(w, "Not found", 404)
		return
	}

	// If trying to get home page with other than GET request, serve and error
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.ServeFile(w, r, "public/build/index.html")
}

func main() {
	flag.Parse()

	dbconn := newHalooDB()
	dbconn.connect()

	rooms := getRooms(dbconn)

	go dbconn.queuePump()

	hub := newHub(dbconn)
	go hub.run()

	http.HandleFunc("/chat", serveHome)

	// Start serving websockets for all rooms
	for _, room := range rooms {
		roomHub := newHub(dbconn)
		go roomHub.run()

		http.HandleFunc("/"+strconv.Itoa(room.ID), func(w http.ResponseWriter, r *http.Request) {
			serveWs(roomHub, w, r)
		})
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// Serve Javascript and CSS files
	fs := http.FileServer(http.Dir("public/build/static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	// Get all rooms and conversations for one user so that they can be displayed in the UI
	http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type UserConversationInfo struct {
			Conversations []User `json:"conversations"`
			Rooms         []Room `json:"rooms"`
		}

		log.Println(r.URL)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		// User ID comes in query parameter user_id
		userIds, ok := r.URL.Query()["user_id"]

		if !ok || len(userIds) < 1 {
			log.Printf("no user_id provided for getting conversations & rooms")
			// TODO: Return JSON stating the error.
		}

		// There should only be one user_id coming from the request
		strID := userIds[0]

		userID, err := strconv.Atoi(strID)
		if err != nil {
			log.Printf("error converting userid to integer: %v", err)
		}
		user := getUser(dbconn, userID)

		log.Printf("User id is: %v", user.ID)

		// Conversations with other users
		conversations := user.getConversations()

		// Rooms user is in
		rooms := user.getRooms()

		log.Printf("rooms: %v", rooms)

		// Bundle user conversations and rooms into one JSON data
		var conversationInfo UserConversationInfo
		conversationInfo.Conversations = conversations
		conversationInfo.Rooms = rooms

		log.Printf("conversation info: %v", conversationInfo)

		conversationJSON, err := json.Marshal(conversationInfo)
		if err != nil {
			log.Printf("error converting conversations to JSON: %v", err)
		}

		// Return JSON for user
		w.Write(conversationJSON)
	})

	http.HandleFunc("/chatlog", func(w http.ResponseWriter, r *http.Request) {
		type ChatlogJSON struct {
			Sender    string `json:"sender"`
			Receiver  string `json:"receiver"`
			Message   string `json:"message"`
			RoomID    int    `json:"room_id"`
			Timestamp int64  `json:"timestamp"`
			Name      string `json:"name"`
		}

		var chatData []ChatlogJSON

		w.Header().Set("Content-Type", "application/json")

		log.Println(r.URL)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		roomID, ok := r.URL.Query()["room_id"]
		if !ok || len(roomID) < 1 {
			userID, ok := r.URL.Query()["user_id"]
			if !ok || len(userID) < 1 {
				log.Printf("no user_id provided for getting conversations & rooms")
				// TODO: Return JSON stating the error.
			}

			receiverID, ok := r.URL.Query()["receiver_id"]
			if !ok || len(receiverID) < 1 {
				log.Printf("no receiver_id provided for getting conversations & rooms")
				// TODO: Return JSON stating the error.
			}

			rows, err := dbconn.connection.Query("SELECT c.sender, c.receiver, c.message, c.timestamp, cu.name FROM chatlog c JOIN chat_users cu ON cu.id = c.receiver WHERE ((sender = $1) AND (receiver = $2)) AND (room_id IS NULL);", userID[0], receiverID[0])
			if err != nil {
				log.Printf("error reading chatlog for user: %v", err)
			}

			defer rows.Close()
			for rows.Next() {
				var cData ChatlogJSON
				if err := rows.Scan(&cData.Sender, &cData.Receiver, &cData.Message, &cData.Timestamp, &cData.Name); err != nil {
					log.Printf("error reading chatlog data %v", err)
				}

				chatData = append(chatData, cData)
			}
		} else {
			rows, err := dbconn.connection.Query("SELECT sender, receiver, message, room_id, timestamp FROM chatlog WHERE room_id = $1", roomID[0])
			if err != nil {
				log.Printf("error reading chatlog for room: %v", err)
			}

			log.Printf("rows: %v", rows)

			defer rows.Close()
			for rows.Next() {
				var cData ChatlogJSON
				if err := rows.Scan(&cData.Sender, &cData.Receiver, &cData.Message, &cData.RoomID, &cData.Timestamp); err != nil {
					log.Printf("error reading chatlog data: %v", err)
				}

				chatData = append(chatData, cData)
			}
		}

		log.Printf("chat data: %v", chatData)

		chatDataJSON, err := json.Marshal(chatData)
		if err != nil {
			log.Printf("error converting json: %v", err)
		}

		w.Write(chatDataJSON)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
