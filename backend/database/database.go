package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var Dbpool *pgxpool.Pool

type RedisClients struct {
	Client0 *redis.Client
	Client1 *redis.Client
}

var RedisAllClients RedisClients

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

func RedisInit(redisURL ...string) error {
	opts0, err := redis.ParseURL(redisURL[0])
	if err != nil {
		return fmt.Errorf("unable to run redis: %v", err)
	}

	client0 := redis.NewClient(opts0)
	ctx := context.Background()

	err = client0.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("unable to ping redis: %v", err)
	}

	RedisAllClients.Client0 = client0
	fmt.Println("Connected to redis 0")

	//-----------

	opts2, err := redis.ParseURL(redisURL[0])
	if err != nil {
		return fmt.Errorf("unable to run redis: %v", err)
	}

	client1 := redis.NewClient(opts2)
	ctx2 := context.Background()

	err = client0.Ping(ctx2).Err()
	if err != nil {
		return fmt.Errorf("unable to ping redis: %v", err)
	}

	RedisAllClients.Client1 = client1
	fmt.Println("Connected to redis 1")
	return nil
}

func RedisClose() {
	if RedisAllClients.Client0 != nil {
		RedisAllClients.Client0.Close()
	}
	if RedisAllClients.Client1 != nil {
		RedisAllClients.Client1.Close()
	}
}
