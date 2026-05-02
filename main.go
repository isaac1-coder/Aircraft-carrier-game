package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Entity struct {
	ID   string  `json:"id"`
	Type string  `json:"type"` 
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Ry   float64 `json:"ry"`
}

var (
	state   = make(map[string]*Entity)
	stateMu sync.Mutex
	clients = make(map[*websocket.Conn]string)
	mu      sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", handleConnections)

	log.Printf("Carrier Server active on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade Error:", err)
		return
	}
	defer ws.Close()

	// Generate unique ID and use it immediately to satisfy the compiler
	playerID := fmt.Sprintf("user_%d", time.Now().UnixNano())
	log.Printf("Player %s joined the deck", playerID)

	mu.Lock()
	clients[ws] = playerID
	mu.Unlock()

	for {
		var msg Entity
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Player %s disconnected", playerID)
			stateMu.Lock()
			delete(state, playerID)
			stateMu.Unlock()
			
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}

		// Update global state
		msg.ID = playerID
		stateMu.Lock()
		state[playerID] = &msg
		
		// Broadcast to all
		payload, _ := json.Marshal(state)
		
		mu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, payload)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
		stateMu.Unlock()
	}
}
