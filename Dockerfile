# Etapa 1: Construção da aplicação
FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

# Compile o aplicativo Go
RUN go mod tidy && go build -o main .

# Etapa 2: Criar a imagem final mínima
FROM alpine:latest

# Copie apenas o binário compilado da etapa anterior
COPY --from=builder /app/main /

# Comando para executar o binário
CMD ["./main"]