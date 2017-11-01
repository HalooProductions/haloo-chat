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
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	dbconn := newHalooDB()
	dbconn.connect()

	rooms := getRooms(dbconn)

	go dbconn.queuePump()

	hub := newHub(dbconn)
	go hub.run()

	http.HandleFunc("/", serveHome)

	for _, room := range rooms {
		roomHub := newHub(dbconn)
		go roomHub.run()

		http.HandleFunc("/"+room.Name, func(w http.ResponseWriter, r *http.Request) {
			serveWs(roomHub, w, r)
		})
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type UserConversationInfo struct {
			conversations []User
			rooms         []Room
		}

		log.Println(r.URL)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		userIds, ok := r.URL.Query()["user_id"]

		if !ok || len(userIds) < 1 {
			log.Printf("no user_id provided for getting conversations & rooms")
			// TODO: Return JSON stating the error.
		}

		strID := userIds[0]

		userID, err := strconv.Atoi(strID)
		if err != nil {
			log.Printf("error converting userid to integer: %v", err)
		}
		user := getUser(dbconn, userID)

		// Conversations with other users
		conversations := user.getConversations(dbconn)

		// Rooms user is in
		rooms := user.getRooms(dbconn)

		var conversationInfo UserConversationInfo
		conversationInfo.conversations = conversations
		conversationInfo.rooms = rooms

		conversationJSON, err := json.Marshal(conversationInfo)
		if err != nil {
			log.Printf("error converting conversations to JSON: %v", err)
		}

		w.Write(conversationJSON)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
