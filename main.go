package main

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxt "gopkg.in/DataDog/dd-trace-go.v1/contrib/jackc/pgx"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	if _, ok := os.LookupEnv("DD_SERVICE"); !ok {
		os.Setenv("DD_SERVICE", "pgx-ddtrace-example")
	}

	if _, ok := os.LookupEnv("DD_ENV"); !ok {
		os.Setenv("DD_ENV", "development")
	}

	tracer.Start(
		tracer.WithAnalytics(true),
	)
	defer tracer.Stop()

	// Start a root span.
	span, ctx := tracer.StartSpanFromContext(context.Background(), "fake.span")
	defer span.Finish()

	dbURLString := os.Getenv("EXAMPLE_DATABASE_URL")
	if dbURLString == "" {
		dbURLString = "postgresql://localhost"
	}

	config, err := pgxpool.ParseConfig(dbURLString)
	if err != nil {
		log.Error("Unable to parse connection string", "err", err)
		os.Exit(1)
	}

	db, err := pgxt.TracedPoolWithConfig(ctx, config)
	if err != nil {
		log.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	for i := 0; i < 5; i++ {
		child := tracer.StartSpan("db.lookup", tracer.ChildOf(span.Context()))
		err := hitDatabase(ctx, db)
		child.Finish(tracer.WithError(err))
	}
}

func hitDatabase(ctx context.Context, db *pgxpool.Pool) error {
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
	err := pgxscan.Select(ctx, db, &result, queryString)

	if err != nil {
		log.Error("select failed", "err", err)
		return err
	}

	log.Info("Result from query was", "result", result[0])
	return nil
}
