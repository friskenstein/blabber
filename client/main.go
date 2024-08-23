package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func() {
		fmt.Println("Closing connection...")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}()

	// Send the user's name to the server
	err = conn.WriteMessage(websocket.TextMessage, []byte(name))
	if err != nil {
		log.Println("Error sending name:", err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("%s\n", message)
		}
	}()

	fmt.Print("You have entered the chat. Type '.quit' to exit.\n")
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == ".quit" {
			fmt.Println("Disconnecting from server...")
			break
		}
		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			log.Println("write:", err)
			return
		}

		select {
		case <-done:
			return
		case <-time.After(time.Second):
		}
	}
}
