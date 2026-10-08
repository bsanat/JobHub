package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"jobhub-ai/backend/internal/database"
)

func main() {
	path := flag.String("sql", "", "reviewed SQL patch path")
	flag.Parse()
	if err := run(*path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Telegram public-import trigger guard applied")
}

func run(path string) error {
	if path == "" || flag.NArg() != 0 || os.Getenv("APP_ENV") != "production" {
		return errors.New("a reviewed SQL patch and APP_ENV=production are required")
	}
	raw := os.Getenv("DATABASE_URL")
	dbURL, err := url.Parse(raw)
	if err != nil || (dbURL.Scheme != "postgresql" && dbURL.Scheme != "postgres") || dbURL.Hostname() == "" || dbURL.User == nil || dbURL.Hostname() != os.Getenv("PUBLIC_IMPORT_DATABASE_HOST") || dbURL.User.Username() != os.Getenv("PUBLIC_IMPORT_DATABASE_USER") {
		return errors.New("production database identity check failed")
	}
	statement, err := os.ReadFile(path)
	if err != nil {
		return errors.New("cannot read reviewed SQL patch")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Connect(ctx, raw, "exec")
	if err != nil {
		return errors.New("database connection failed")
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, string(statement)); err != nil {
		return errors.New("Telegram trigger patch failed")
	}
	return nil
}
