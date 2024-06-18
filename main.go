package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	databaseURL string
	secretKey   string
	port        string
)

const (
	serverCode   = http.StatusInternalServerError
	unauthorized = http.StatusUnauthorized
	statusOK     = http.StatusOK
	badCode      = http.StatusBadRequest
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file")
	}
	databaseURL = os.Getenv("DATABASE_URL")
	secretKey = os.Getenv("SECRET_KEY")
	port = os.Getenv("PORT")
}

func main() {
	if err := InitDB(); err != nil {
		log.Printf("Failed to initialize database: %v", err)
	}

	defer Pool.Close()
	// err := createSchema()
	// if err != nil {
	// 	log.Printf("Unable to create schema: %v\n", err)
	// }

	routes()
}
