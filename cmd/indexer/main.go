package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "db-url", "", "url to connect to postgres db")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite3", "oshistats.db?_journal_mode=WAL")
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	store, err := vtubers.CreateStore(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	pgDB, err := sqlx.Open("pgx", dbURL)
	if err != nil {
		log.Panicln(err)
	}
	defer pgDB.Close()

	logRepository := logs.NewRepository(pgDB)

	ls, err := logRepository.GetAll(ctx)
	if err != nil {
		log.Panicln(err)
	}
	defer ls.Close()

	repo, err := index.CreateIndexedVideoRepository(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	indexer := index.NewIndexer(store, logRepository, repo)
	err = indexer.Index(ctx)
	if err != nil {
		log.Panicln(err)
	}
}
