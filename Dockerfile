# Etapa 1: Construção da aplicação
FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

# Compile o aplicativo Go
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Etapa 2: Criar a imagem final mínima
FROM alpine:latest

# Copie apenas o binário compilado da etapa anterior
COPY --from=builder /app/server /

# Expor a porta correta
EXPOSE 8080

# Definir variável de ambiente para Cloud Run
ENV PORT=8080

# Comando para executar o binário
CMD ["./server"]