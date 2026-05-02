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
	ID    string  `json:"id"`
	Type  string  `json:"type"` 
	X, Y, Z float64 `json:"x"`
	Rx, Ry, Rz float64 `json:"rx"`
	Health int    `json:"hp"`
}

var (
	state   = make(map[string]*Entity)
	stateMu sync.Mutex
	clients = make(map[*websocket.Conn]string)
	clientsMu sync.Mutex
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // Render standard port
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", handleConnections)

	// AI/Bot Logic Loop
	go func() {
		for {
			stateMu.Lock()
			// Update bots or clean up inactive players
			stateMu.Unlock()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	log.Println("Warship Server running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// FIXED: Generating ID and using it immediately
	id := fmt.Sprintf("p%d", time.Now().UnixNano())
	
	clientsMu.Lock()
	clients[ws] = id
	clientsMu.Unlock()

	log.Printf("Player %s connected", id)

	for {
		var msg Entity
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Player %s disconnected", id)
			stateMu.Lock()
			delete(state, id)
			stateMu.Unlock()
			
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			break
		}

		// Update state with the player's ID
		msg.ID = id
		stateMu.Lock()
		state[id] = &msg
		
		// Broadcast state to all connected clients
		payload, _ := json.Marshal(state)
		clientsMu.Lock()
		for conn := range clients {
			conn.WriteMessage(websocket.TextMessage, payload)
		}
		clientsMu.Unlock()
		stateMu.Unlock()
	}
}
