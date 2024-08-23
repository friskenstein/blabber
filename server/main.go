package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	name string
}

var clients = make(map[*websocket.Conn]*Client)
var broadcast = make(chan string)

func newConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		delete(clients, ws)
		ws.Close()
	}()

	_, name, err := ws.ReadMessage()
	if err != nil {
		log.Printf("Failed to read name: %v", err)
		return
	}

	client := &Client{conn: ws, name: string(name)}
	clients[ws] = client

	log.Printf("%s joined the chat", client.name)

	broadcast <- client.name + " joined the chat"

	sendToOne(client.conn, "Welcome, "+client.name+"!\nActive users: "+strings.Join(getUserList(), ", "))

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("%s disconnected: %v", client.name, err)
			broadcast <- client.name + " left the chat"
			break
		}
		broadcast <- client.name + ": " + string(msg)
	}
}

func getUserList() []string {
	var userList []string
	for _, client := range clients {
		userList = append(userList, client.name)
	}
	return userList
}

func sendToOne(conn *websocket.Conn, msg string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Printf("Error sending message: %v", err)
		conn.Close()
		delete(clients, conn)
	}
}

func sendToAll(msg string) {
	for _, client := range clients {
		sendToOne(client.conn, msg)
	}
}

func main() {
	http.HandleFunc("/ws", newConnection)
	go func () {
		for {
			sendToAll(<-broadcast)
		}
	}()

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
