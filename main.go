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
	port := os.Getenv("PORT") // Cloud Run define a porta via variável de ambiente
	if port == "" {
		port = "8080" // Define 8080 como padrão caso a variável esteja vazia
	}

	http.HandleFunc("/", handler)

	log.Printf("Server is listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
