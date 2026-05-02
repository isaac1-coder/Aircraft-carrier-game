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

type GameState struct {
	ID   string  `json:"id"`
	Type string  `json:"type"` // "player", "missile", "jet"
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	RotY float64 `json:"ry"`
}

var (
	players   = make(map[string]*GameState)
	playersMu sync.Mutex
	clients   = make(map[*websocket.Conn]string)
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "10000" }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", handleConnections)

	log.Println("Battlefield Server Online on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	id := fmt.Sprintf("u_%d", time.Now().UnixNano())
	clients[ws] = id

	for {
		var msg GameState
		err := ws.ReadJSON(&msg)
		if err != nil {
			playersMu.Lock()
			delete(players, id)
			playersMu.Unlock()
			break
		}
		msg.ID = id
		playersMu.Lock()
		players[id] = &msg
		data, _ := json.Marshal(players)
		for conn := range clients {
			conn.WriteMessage(websocket.TextMessage, data)
		}
		playersMu.Unlock()
	}
}
