package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn := getConn()

	// err = createSchema(conn)
	// if err != nil {
	//     log.Fatalf("Unable to create schema: %v\n", err)
	// }

	routes(conn)
}

func getConn() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Connected to database")
	}

	defer conn.Close(context.Background())
	return conn
}
