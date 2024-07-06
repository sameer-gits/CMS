package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Dbpool *pgxpool.Pool

func DbInit(dbURL string) error {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 5
	config.MaxConnIdleTime = 60

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Use the Ping method to check if the connection is successful
	err = pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to connect to the database: %v", err)
	}

	Dbpool = pool
	fmt.Println("Connected to database")
	return nil
}

func DbClose() {
	if Dbpool != nil {
		Dbpool.Close()
	}
}
