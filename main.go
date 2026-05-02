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
	ID   string  `json:"id"`
	Type string  `json:"type"` // player, bot, missile
	X, Y, Z float64 `json:"x"`
	Ry   float64 `json:"ry"`
	Diff string  `json:"diff"` // recruit, hardened, veteran
}

var (
	entities = make(map[string]*Entity)
	mu       sync.Mutex
	clients  = make(map[*websocket.Conn]string)
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "10000" }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", handleConnections)

	// Bot AI Loop
	go func() {
		for {
			mu.Lock()
			for id, ent := range entities {
				if ent.Type == "bot" {
					speed := 0.05
					if ent.Diff == "Veteran" { speed = 0.2 }
					ent.Z += speed // Bots move forward
					if ent.Z > 200 { ent.Z = -200 }
				}
			}
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	log.Println("Navy Ops Server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	id := fmt.Sprintf("u%d", time.Now().UnixNano())
	clients[ws] = id

	for {
		var msg Entity
		err := ws.ReadJSON(&msg)
		if err != nil {
			mu.Lock()
			delete(entities, id)
			mu.Unlock()
			break
		}
		msg.ID = id
		mu.Lock()
		entities[id] = &msg
		payload, _ := json.Marshal(entities)
		for c := range clients {
			c.WriteMessage(websocket.TextMessage, payload)
		}
		mu.Unlock()
	}
}
