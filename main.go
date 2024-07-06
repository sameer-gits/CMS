package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sameer-gits/CMS/database"
)

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file")
	}
}

func main() {
	err := database.DbInit(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.DbClose()
	routes()
}
