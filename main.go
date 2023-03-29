package main

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURLString := os.Getenv("EXAMPLE_DATABASE_URL")
	if dbURLString == "" {
		dbURLString = "postgresql://localhost"
	}

	queryString := `SELECT
clock_timestamp() as time1,
pg_sleep(.5) as sleep1,
clock_timestamp() as time2,
pg_sleep(.5) as sleep2,
clock_timestamp() as time3;`

	result := []struct {
		Time1  time.Time
		Sleep1 string
		Time2  time.Time
		Sleep2 string
		Time3  time.Time
	}{}

	ctx := context.Background()
	config, err := pgxpool.ParseConfig(dbURLString)
	if err != nil {
		log.Error("Unable to parse connection string", "err", err)
		os.Exit(1)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	err = pgxscan.Select(ctx, db, &result, queryString)

	if err != nil {
		log.Error("select failed", "err", err)
		os.Exit(1)
	}

	log.Info("Result from query was", "result", result[0])
}
