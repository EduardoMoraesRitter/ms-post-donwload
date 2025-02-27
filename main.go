package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Porta padrão para Cloud Run
	}

	http.HandleFunc("/", handler)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
