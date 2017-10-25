// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
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

	rooms := dbconn.getRooms()

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

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
