package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool

type InitOpts struct {
	EnvPath string
}

func InitDb(opts *InitOpts) error {
	var err error
	if opts == nil {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(opts.EnvPath)
	}
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %w, err")
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	return nil
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
	if pool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}
	return pool.Query(ctx, query, args...)
}

func QueryRow(ctx context.Context, query string, args ...interface{}) (pgx.Row, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}
	return pool.QueryRow(ctx, query, args...), nil
}

func Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	if pool == nil {
		return pgconn.CommandTag{}, fmt.Errorf("database pool is not initialized")
	}
	return pool.Exec(ctx, query, args...)
}
