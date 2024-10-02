package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool

func InitDb() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	pool, err = pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
}

func CloseDb() {
	if pool != nil {
		pool.Close()
	}
}

func Pool() *pgxpool.Pool {
	return pool
}

func Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return pool.Query(ctx, query, args...)
}

func QueryRow(ctx context.Context, query string, args ...interface{}) (pgx.Row, error) {
	return pool.QueryRow(ctx, query, args...), nil
}

func Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return pool.Exec(ctx, query, args...)
}
