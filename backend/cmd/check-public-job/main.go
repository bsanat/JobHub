package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"jobhub-ai/backend/internal/database"
	"jobhub-ai/backend/internal/jobs"
)

func main() {
	id := flag.String("id", "", "public vacancy UUID to read")
	flag.Parse()
	if *id == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: check-public-job --id <uuid>")
		os.Exit(2)
	}
	raw := os.Getenv("DATABASE_URL")
	dbURL, err := url.Parse(raw)
	if err != nil || dbURL.Hostname() == "" || dbURL.User == nil || dbURL.Hostname() != os.Getenv("PUBLIC_IMPORT_DATABASE_HOST") || dbURL.User.Username() != os.Getenv("PUBLIC_IMPORT_DATABASE_USER") {
		fmt.Fprintln(os.Stderr, "production database identity check failed")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Connect(ctx, raw, "exec")
	if err != nil {
		fmt.Fprintln(os.Stderr, "database connection failed")
		os.Exit(1)
	}
	defer pool.Close()
	_, err = jobs.NewPostgresStore(pool, true).Get(ctx, *id)
	if err == nil {
		fmt.Println("public job detail query succeeded")
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		fmt.Fprintf(os.Stderr, "public job detail SQL error: code=%s message=%s\n", pgErr.Code, pgErr.Message)
	} else {
		fmt.Fprintf(os.Stderr, "public job detail read error: %v\n", err)
	}
	os.Exit(1)
}
