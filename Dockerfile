# Etapa 1: Construção da aplicação
FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

# Compilar um binário estático para evitar problemas no Alpine
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Etapa 2: Criar a imagem final mínima
FROM alpine:latest

WORKDIR /

# Copiar apenas o binário compilado
COPY --from=builder /app/server /

# Expor a porta correta
EXPOSE 8080

# Definir variável de ambiente para Cloud Run
ENV PORT=8080

# Iniciar o servidor Go
CMD ["/server"]
