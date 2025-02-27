package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

// Estrutura do JSON recebido no POST
type MediaRequest struct {
	Channel   string `json:"channel"`
	CreatorID int    `json:"creator_id"`
	MediaURL  string `json:"media_url"`
}

// Configuração do Storage (substitua pelo seu bucket)
var bucketName = "smart_match_creators_test"

const (
	maxVideoSize = 100 * 1024 * 1024 // 100MB - Cancela o download se ultrapassar
	minifyLimit  = 10 * 1024 * 1024  // 10MB - Faz compressão se ultrapassar
)

// 📌 **1. Manipulador HTTP para receber o JSON**
func handleMediaDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica o JSON recebido
	var requestData MediaRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Verifica se os campos necessários estão presentes
	if requestData.MediaURL == "" || requestData.Channel == "" || requestData.CreatorID == 0 {
		http.Error(w, "Campos obrigatórios ausentes", http.StatusBadRequest)
		return
	}

	log.Printf("Recebendo vídeo do canal: %s, Criador ID: %d", requestData.Channel, requestData.CreatorID)

	// 📌 Obtém metadados do arquivo antes de baixar
	shouldDownload, contentType, fileSize := checkFileMetadata(requestData.MediaURL)
	if !shouldDownload {
		http.Error(w, "Arquivo muito grande ou erro ao verificar metadados", http.StatusForbidden)
		return
	}

	// 📌 Faz o download do vídeo
	filePath := downloadFile(requestData.MediaURL, contentType)
	if filePath == "" {
		http.Error(w, "Erro ao baixar o arquivo", http.StatusInternalServerError)
		return
	}

	// 📌 Minifica se necessário
	finalFilePath := minifyVideo(filePath, fileSize)

	// 📌 Upload para Google Cloud Storage
	client, err := storage.NewClient(context.Background())
	if err != nil {
		http.Error(w, "Erro ao conectar ao Storage", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Define o caminho baseado no canal e creator_id
	storagePath := fmt.Sprintf("%s/%d/%s", requestData.Channel, requestData.CreatorID, filepath.Base(finalFilePath))

	storageURI, err := uploadToStorage(client, finalFilePath, storagePath)
	if err != nil {
		http.Error(w, "Erro no upload", http.StatusInternalServerError)
		return
	}

	// Responde com sucesso e o caminho do arquivo no Storage
	response := map[string]string{
		"message":   "Upload concluído",
		"file_path": storageURI,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 📌 **2. Obtém Metadados do Arquivo**
func checkFileMetadata(url string) (bool, string, int) {
	resp, err := http.Head(url)
	if err != nil {
		log.Println("Erro ao fazer HEAD request:", err)
		return false, "", 0
	}
	defer resp.Body.Close()

	contentLengthStr := resp.Header.Get("Content-Length")
	contentType := resp.Header.Get("Content-Type")

	if contentLengthStr == "" {
		log.Println("Não foi possível obter o tamanho do arquivo.")
		return false, "", 0
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		log.Println("Erro ao converter Content-Length:", err)
		return false, "", 0
	}

	// Se for maior que 100MB, cancela o download
	if contentLength > maxVideoSize {
		log.Println("Arquivo maior que 100MB. Download cancelado.")
		return false, "", 0
	}

	return true, contentType, contentLength
}

// 📌 **3. Faz Download do Arquivo**
func downloadFile(url, contentType string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Erro ao baixar o arquivo:", err)
		return ""
	}
	defer resp.Body.Close()

	ext := getExtension(contentType)
	fileName := fmt.Sprintf("video_%d%s", time.Now().UnixNano(), ext)
	file, err := os.Create(fileName)
	if err != nil {
		log.Println("Erro ao criar arquivo:", err)
		return ""
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println("Erro ao salvar arquivo:", err)
		return ""
	}

	return fileName
}

// 📌 **4. Minifica o Vídeo se Necessário**
func minifyVideo(inputFile string, fileSize int) string {
	if fileSize < minifyLimit {
		log.Println("Arquivo menor que 10MB, minificação não necessária.")
		return inputFile
	}

	outputFile := fmt.Sprintf("%s_minified.mp4", inputFile[:len(inputFile)-len(filepath.Ext(inputFile))])
	log.Println("Iniciando minificação com FFmpeg...")

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputFile,
		"-vf", "scale=trunc(iw*0.5/2)*2:trunc(ih*0.5/2)*2,fps=30",
		"-c:v", "libx264",
		"-preset", "slow",
		"-y", outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Println("Erro ao executar FFmpeg:", stderr.String())
		return inputFile
	}

	return outputFile
}

// 📌 **5. Faz Upload para o Google Cloud Storage**
func uploadToStorage(client *storage.Client, filePath, storagePath string) (string, error) {
	ctx := context.Background()
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	wc := client.Bucket(bucketName).Object(storagePath).NewWriter(ctx)
	_, err = io.Copy(wc, file)
	if err != nil {
		return "", fmt.Errorf("erro ao escrever no Storage: %v", err)
	}

	err = wc.Close()
	if err != nil {
		return "", fmt.Errorf("erro ao finalizar upload no Storage: %v", err)
	}

	return fmt.Sprintf("gs://%s/%s", bucketName, storagePath), nil
}

// 📌 **6. Infere a Extensão pelo Content-Type**
func getExtension(contentType string) string {
	mimeTypes := map[string]string{
		"video/mp4":        ".mp4",
		"video/x-matroska": ".mkv",
		"video/avi":        ".avi",
		"video/quicktime":  ".mov",
		"video/webm":       ".webm",
	}
	if ext, exists := mimeTypes[contentType]; exists {
		return ext
	}
	return ".mp4"
}

// 📌 **7. Inicia o Servidor**
func main() {
	http.HandleFunc("/upload", handleMediaDownload)
	fmt.Println("Servidor rodando em http://localhost:8080/upload")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
