# Usar uma imagem leve do Go
FROM golang:1.21 as builder

# Criar diretório de trabalho
WORKDIR /app

# Copiar arquivos
COPY . .

# Baixar dependências e compilar
RUN go mod init hello && go mod tidy && go build -o server

# Criar imagem final
FROM alpine:latest

WORKDIR /

# Copiar o binário compilado
COPY --from=builder /app/server .

# Expor a porta
EXPOSE 8080

# Executar o binário
CMD ["/server"]
