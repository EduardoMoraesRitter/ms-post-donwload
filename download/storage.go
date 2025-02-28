package download

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"postDownload/configs"

	"cloud.google.com/go/storage"
)

// 📌 **1. Faz Upload do Arquivo para o Google Cloud Storage**
func UploadToStorage(client *storage.Client, finalFilePath, path string) (string, error) {

	w := client.Bucket(configs.Env.BucketSmartMatchCreators).Object(path).NewWriter(configs.Env.Ctx)

	// **Faz o upload do arquivo para o Storage**
	fileData, err := os.ReadFile(finalFilePath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	//upload
	_, err = w.Write(fileData)
	if err != nil {
		return "", fmt.Errorf("erro ao escrever no Storage: %v", err)
	}
	w.Close()
	defer os.Remove(finalFilePath)

	uri := fmt.Sprintf("gs://%s/%s", configs.Env.BucketSmartMatchCreators, path)
	return uri, nil
}

// 📌 **2. Verifica se o Arquivo já Existe no Storage**
func FileExistsInStorage(client *storage.Client, path string) (bool, string, error) {
	bucket := client.Bucket(configs.Env.BucketSmartMatchCreators)
	obj := bucket.Object(path)

	_, err := obj.Attrs(configs.Env.Ctx)
	if err == nil {
		ext := filepath.Ext(path)
		if ext == "" {
			log.Printf("Arquivo sem extensão detectado: %s. Deletando do Storage...", path)
			if err := obj.Delete(configs.Env.Ctx); err != nil {
				log.Printf("Erro ao deletar arquivo sem extensão: %v", err)
				return false, "", fmt.Errorf("erro ao deletar arquivo sem extensão: %v", err)
			}
			return false, "", nil
		}

		uri := fmt.Sprintf("gs://%s/%s", configs.Env.BucketSmartMatchCreators, path)
		return true, uri, nil
	}

	if err == storage.ErrObjectNotExist {
		return false, "", nil
	}

	return false, "", fmt.Errorf("erro ao verificar existência do arquivo no Storage: %v", err)
}

// 📌 **3. Faz Download do Arquivo**
func DownloadFile(fileURL, fileName string) string {

	resp, err := http.Get(fileURL)
	if err != nil {
		log.Println("Erro ao baixar o arquivo:", err)
		return ""
	}
	defer resp.Body.Close()

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

// 📌 **4. Minificação do Vídeo**
func MinifyVideo(inputFile string) string {
	log.Println("Criando arquivo temporário de entrada")
	input, err := os.Create(fmt.Sprintf("input_%d", time.Now().UnixNano()))
	if err != nil {
		log.Println("Erro ao criar arquivo de entrada:", err)
		return inputFile
	}
	defer os.Remove(input.Name())

	log.Println("Escrevendo no arquivo temporário")
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Println("Erro ao ler arquivo:", err)
		return inputFile
	}

	if _, err := input.Write(data); err != nil {
		log.Println("Erro ao escrever no arquivo temporário:", err)
		return inputFile
	}
	input.Close()

	log.Println("Criando arquivo temporário de saída")
	//output := fmt.Sprintf("%s_minify.mp4", inputFile)
	//defer os.Remove(output)

	log.Println("Calculando fator de escala")
	fileSizeMB := float64(len(data)) / (1024 * 1024)
	scaleFactor := 1.0 - (math.Log10(fileSizeMB) / 3.8)

	if scaleFactor < 0.2 {
		scaleFactor = 0.2
	} else if scaleFactor > 0.9 {
		scaleFactor = 0.9
	}
	scaleFactor = math.Round(scaleFactor*100) / 100

	cmd := exec.Command(
		"ffmpeg",
		"-i", input.Name(),
		"-vf", fmt.Sprintf("scale=trunc(iw*%.2f/2)*2:trunc(ih*%.2f/2)*2,fps=30", scaleFactor, scaleFactor),
		"-c:v", "libx264",
		"-preset", "slow",
		inputFile,
		"-y",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Println("Executando compressão com FFmpeg", "scale_factor", scaleFactor)
	if err := cmd.Run(); err != nil {
		log.Println("Erro ao executar FFmpeg:", stderr.String())
		log.Println("Erro FFmpeg:", err)
		return inputFile
	}

	log.Println("Vídeo minificado:", inputFile)
	return inputFile
}
