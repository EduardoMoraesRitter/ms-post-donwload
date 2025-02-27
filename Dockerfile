# Etapa 1: Builder
FROM golang:1.22.3-alpine AS builder

# Instale pacotes necessários para build
RUN apk add --no-cache git

WORKDIR /app

# Copie o código fonte e outros arquivos necessários para o build
COPY . .

# Compile o aplicativo Go
RUN go build -o main ./main.go

# Etapa 2: Imagem final
FROM alpine:latest

# Copie apenas o binário compilado da etapa anterior
COPY --from=builder /app/main .

# Defina as variáveis de ambiente
ENV PORT=8080
ENV GIN_MODE=release
ENV ENV=production

# Exponha a porta
EXPOSE ${PORT}

# Comando para executar o binário
CMD ["./main"]
