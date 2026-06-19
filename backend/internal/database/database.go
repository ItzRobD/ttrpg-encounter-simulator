package database

import (
	"context"
	"fmt"
	"net/url"
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
	// A missing .env file is not fatal: in production the DB_* vars are
	// supplied by the environment directly. Only surface other load errors.
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error loading env file: %w", err)
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// URL-escape credentials so passwords containing reserved characters
	// (@, :, /, etc.) don't corrupt the connection string.
	connString := fmt.Sprintf("postgres://%s@%s:%s/%s",
		url.UserPassword(user, password).String(),
		host, port, url.PathEscape(dbname))
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %w", err)
	}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
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
