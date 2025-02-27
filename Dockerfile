# Etapa 1: Construção da aplicação
FROM golang:1.22 AS builder 

# Definir diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos do projeto
COPY . .

# Baixar dependências e compilar o binário
RUN go mod tidy && go build -o server .

# Etapa 2: Criar a imagem final minimalista
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copiar o binário da etapa anterior
COPY --from=builder /app/server /

# Expor a porta usada pelo Cloud Run
EXPOSE 8080

# Definir comando de inicialização do container
CMD ["/server"]
