package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	connections map[string]*websocket.Conn = make(map[string]*websocket.Conn)
	keys        map[string]string          = make(map[string]string)
)

type Message struct {
	Type    string `json:"type"`
	Id      string `json:"id"`
	Message string `json:"message"`
}

func startWebServer() {
	http.HandleFunc("/ws", handleWebsockets)
	http.HandleFunc("/", handleRoot)
	http.ListenAndServe("[::]:5001", nil)
	wg.Done()
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		// If there was a problem, stop
		return
	}

	// Generate random id for users
	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		return
	}

	tmpl.ExecuteTemplate(w, "index", struct {
		Id string
	}{
		Id: base64.URLEncoding.EncodeToString(bytes),
	})
}

func handleWebsockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Save the id to help with deleting the connection later
	var id string

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			// Only delete the connection and keys if we ever saved it with the id
			if id != "" {
				delete(connections, id)
				delete(keys, id)
			}
			break
		}

		// Save connection and id if they're not already set
		if id == "" {
			id = message.Id
			connections[id] = conn
		}

		switch message.Type {
		case "SAVE-KEY":
			keys[id] = message.Message
		}
	}
}
