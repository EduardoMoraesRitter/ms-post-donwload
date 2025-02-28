package configs

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Variables struct {
	Port                     int    `envconfig:"PORT" required:"true" default:"8080"`
	BucketSmartMatchCreators string `envconfig:"BUCKET_SMART_MATCH_CREATORS" required:"true" default:"smart_match_creators_test"`
	Ctx                      context.Context
}

// Estrutura do JSON recebido no POST
type MediaRequest struct {
	Channel   string `json:"channel"`
	CreatorID int    `json:"creator_id"`
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
