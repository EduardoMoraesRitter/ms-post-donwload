# Etapa 1: Construção da aplicação
FROM golang:1.23-alpine AS builder

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos necessários
COPY go.mod go.sum ./
RUN go mod tidy

# Copiar todo o código-fonte
COPY . .

# Compilar o binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Etapa 2: Criar a imagem final mínima com Alpine
FROM alpine:latest

# Definir diretório de trabalho no contêiner
WORKDIR /root/

# Copiar apenas o binário compilado
COPY --from=builder /app/main .

# Expor a porta correta
EXPOSE 8080

# Comando para executar o binário
CMD ["./main"]
