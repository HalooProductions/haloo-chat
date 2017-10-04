// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8000", "http service address")

// User model
type User struct {
	name     string
	email    string
	password string
}

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
	go dbconn.queuePump()

	rows, err := dbconn.connection.Query("SELECT name, email, password FROM users")
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
		fmt.Println("Error connecting to database")
	}

	defer rows.Close()
	for rows.Next() {
		var userName, email, password string
		if err := rows.Scan(&userName, &email, &password); err != nil {
			log.Fatal(err)
		}
		fmt.Println(userName)
	}

	hub := newHub(dbconn)
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
