package data

import (
	"fmt"
	"log"
	"time"

	"postDownload/configs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 📌 **Conectar ao MongoDB**
func connectMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(configs.Env.MongoURI)
	client, err := mongo.Connect(configs.Env.Ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MongoDB: %v", err)
	}
	return client, nil
}

// 📌 **Atualizar Documento para Incluir `uri` do Storage**
func UpdateCreatorURI(post_id, media_uri string) error {
	client, err := connectMongoDB()
	if err != nil {
		return err
	}
	defer client.Disconnect(configs.Env.Ctx)

	collection := client.Database(configs.Env.MongoDatabase).Collection(configs.Env.MongoCollection)

	// 📌 Filtro: Localiza o documento pelo `creator_id`
	filter := bson.M{"post_id": post_id}

	// 📌 Atualizar documento: Adicionar `media_uri`
	update := bson.M{
		"$set": bson.M{
			"media_uri":  media_uri,
			"updated_at": time.Now(), // Atualiza timestamp
		},
	}

	// Executa a atualização
	result, err := collection.UpdateOne(configs.Env.Ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar criador: %v", err)
	}

	// Verifica se algum documento foi modificado
	if result.MatchedCount == 0 {
		log.Printf("Nenhum documento encontrado para post_id %v", post_id)
	} else {
		log.Printf("URI atualizado com sucesso para post_id %v", post_id)
	}

	return nil
}
