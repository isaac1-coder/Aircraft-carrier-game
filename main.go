package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/", fs)

	// WASM files must be served with the correct MIME type
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
