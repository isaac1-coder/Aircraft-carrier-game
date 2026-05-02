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

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type Entity struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"` // "player", "bot", "missile", "plane"
	X, Y, Z float64 `json:"x"`
	Rx, Ry, Rz float64 `json:"rx"`
	Health int    `json:"hp"`
}

var (
	state   = make(map[string]*Entity)
	stateMu sync.Mutex
	clients = make(map[*websocket.Conn]string)
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", handleConnections)

	// AI Loop - Simple but efficient
	go func() {
		for {
			stateMu.Lock()
			for id, ent := range state {
				if ent.Type == "bot" {
					ent.X += 0.1 // Basic movement logic
				}
			}
			stateMu.Unlock()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	fmt.Println("Warship Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	id := fmt.Sprintf("p%d", time.Now().UnixNano())
	clients[ws] = id

	for {
		var msg Entity
		err := ws.ReadJSON(&msg)
		if err != nil {
			stateMu.Lock()
			delete(state, id)
			stateMu.Unlock()
			break
		}
		msg.ID = id
		stateMu.Lock()
		state[id] = &msg
		
		// Broadcast state to all
		broadcast, _ := json.Marshal(state)
		for conn := range clients {
			conn.WriteMessage(websocket.TextMessage, broadcast)
		}
		stateMu.Unlock()
	}
}
