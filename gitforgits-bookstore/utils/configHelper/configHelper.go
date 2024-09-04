package configHelper

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configuration struct {
	ServerAddress string
	DbUser        string
	DbPassword    string
	DbHost        string
	DbName        string
}

func New() Configuration {
	c := Configuration{}
	c.LoadConfiguration()
	return c
}

func (cfg Configuration) LoadConfiguration() {

	//Load .env file during local development
	err := godotenv.Load(".env")
	if err == nil {
		log.Println("Error loading .env file, assuming production environment with OS level environment variables.")
	}

	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	cfg.DbUser = os.Getenv("DB_USER")
	cfg.DbPassword = os.Getenv("DB_PASSWORD")
	cfg.DbHost = os.Getenv("DB_HOST")
	cfg.DbName = os.Getenv("DB_NAME")
}
