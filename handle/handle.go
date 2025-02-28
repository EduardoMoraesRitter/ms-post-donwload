package handle

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"postDownload/configs"
	"postDownload/data"
	"postDownload/download"

	"cloud.google.com/go/storage"
)

// 📌 **Manipulador HTTP para receber o JSON**
func HandleMediaDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var requestData configs.MediaRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if requestData.MediaURL == "" || requestData.Channel == "" || requestData.CreatorID == 0 || requestData.PostID == "" {
		http.Error(w, "Campos obrigatórios ausentes", http.StatusBadRequest)
		return
	}

	log.Printf("Recebendo vídeo do canal: %s, Criador ID: %d", requestData.Channel, requestData.CreatorID)

	client, err := storage.NewClient(configs.Env.Ctx)
	if err != nil {
		http.Error(w, "Erro ao conectar ao Storage", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// **Extrai o nome do arquivo corretamente**
	fileName := filepath.Base(strings.Split(requestData.MediaURL, "?")[0])

	// **Define o caminho do arquivo no Storage**
	storagePath := fmt.Sprintf("%s/%d/%s", requestData.Channel, requestData.CreatorID, fileName)

	// **Verifica se o arquivo já existe no Storage**
	exists, uri, err := download.FileExistsInStorage(client, storagePath)
	if err != nil {
		http.Error(w, "Erro ao verificar o arquivo no Storage", http.StatusInternalServerError)
		return
	}
	if exists {
		log.Println("Arquivo já existe no Storage:", uri)

		// 📌 **Atualiza o MongoDB com a URI do arquivo já existente**
		if err := data.UpdateCreatorURI(requestData.PostID, uri); err != nil {
			log.Println("Erro ao atualizar MongoDB:", err)
			http.Error(w, "Erro ao atualizar MongoDB", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Arquivo já existe e foi atualizado no MongoDB", "file_path": uri})
		return
	}

	// **Faz o download do arquivo**
	filePath := download.DownloadFile(requestData.MediaURL, fileName)
	if filePath == "" {
		http.Error(w, "Erro ao baixar o arquivo", http.StatusInternalServerError)
		return
	}

	// **Minifica o arquivo se necessário**
	finalFilePath := download.MinifyVideo(filePath)

	// **Faz o upload do arquivo para o Storage**
	storageURI, err := download.UploadToStorage(client, finalFilePath, storagePath)
	if err != nil {
		http.Error(w, "Erro no upload", http.StatusInternalServerError)
		return
	}

	// 📌 **Atualiza o MongoDB com a URI do novo arquivo**
	if err := data.UpdateCreatorURI(requestData.PostID, storageURI); err != nil {
		log.Println("Erro ao atualizar MongoDB:", err)
		http.Error(w, "Erro ao atualizar MongoDB", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message":   "Upload concluído e URI atualizada no MongoDB",
		"file_path": storageURI,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
