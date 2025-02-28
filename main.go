package main

import (
	"fmt"
	"log"
	"net/http"

	"postDownload/configs"
	"postDownload/handle"
)

func main() {
	// 📌 Inicializa variáveis de ambiente
	configs.Init()

	// 📌 Configura rota do upload
	http.HandleFunc("/upload", handle.HandleMediaDownload)

	// 📌 Inicia o servidor HTTP
	port := configs.Env.Port
	log.Printf("Servidor rodando na porta %d 🚀", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
