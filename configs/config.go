package configs

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Variables struct {
	ProjectID                string `envconfig:"PROJECT_ID" required:"true" default:"brandlovrs-develop"`
	Location                 string `envconfig:"LOCATION" required:"true" default:"us-central1"`
	Port                     int    `envconfig:"PORT" required:"true" default:"8080"`
	BucketSmartMatchCreators string `envconfig:"BUCKET_SMART_MATCH_CREATORS" required:"true" default:"smart_match_creators_test"`
	MongoURI                 string `envconfig:"MONGO_URI" required:"true" default:"mongodb+srv://bruno12leonel:6dQDwpKYWmxyCfMe@cluster-teste.czigp.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"`
	MongoDatabase            string `envconfig:"MONGO_DATABASE" required:"true" default:"sample_mflix"`
	MongoCollection          string `envconfig:"MONGO_COLLECTION" required:"true" default:"instagram-posts"`

	Ctx context.Context
}

// Estrutura do JSON recebido no POST
type MediaRequest struct {
	Channel   string `json:"channel"`
	CreatorID int    `json:"creator_id"`
	PostID    string `json:"post_id"`
	MediaURL  string `json:"media_url"`
}

var Env Variables

func Init() {
	// Load variables from .env file, if available
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found, default environment variables will be used")
	}

	// Process environment variables into the Variables struct
	err = envconfig.Process("", &Env)
	if err != nil {
		log.Fatalf("Error processing environment variables: %v", err)
	}

	// Set up the default context
	Env.Ctx = context.Background()

}
