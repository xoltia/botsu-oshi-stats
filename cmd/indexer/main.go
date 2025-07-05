package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
	_ "modernc.org/sqlite"
)

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "db-url", "", "url to connect to postgres db")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite", "oshistats.db")
	if err != nil {
		log.Panicln(err)
	}

	store, err := vtubers.CreateStore(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	pgDB, err := sqlx.Open("pgx", dbURL)
	if err != nil {
		log.Panicln(err)
	}

	logRepository := logs.NewRepository(pgDB)

	ls, err := logRepository.GetAll(ctx)
	if err != nil {
		log.Panicln(err)
	}
	defer ls.Close()

	videoVTuberRepository, err := index.CreateVideoVTuberRepository(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	indexer := index.NewIndexer(store, logRepository, videoVTuberRepository)
	err = indexer.Index(ctx)
	if err != nil {
		log.Panicln(err)
	}
}
