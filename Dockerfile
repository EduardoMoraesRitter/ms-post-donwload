# Etapa 1: Construção da aplicação
FROM golang:1.23 AS builder

# Instale pacotes necessários para build
RUN apk add --no-cache git

WORKDIR /app

COPY . .

# Compile o aplicativo Go
RUN go build -o main ./main.go

# Etapa 2: Criar a imagem final mínima
FROM alpine:latest

# Copie apenas o binário compilado da etapa anterior
COPY --from=builder /app/main .

# Expor a porta correta
EXPOSE 8080

# Definir variável de ambiente para Cloud Run
ENV PORT=8080

# Comando para executar o binário
CMD ["./main"]