package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port       string
	DBName     string
	DeviceName string
}

func NewAppConfig() *AppConfig {
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
	return &AppConfig{
		Port:       os.Getenv("PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DeviceName: os.Getenv("DEVICE_NAME"),
	}
}
