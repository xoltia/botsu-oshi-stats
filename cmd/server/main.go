package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server"
)

func main() {
	var (
		addr  string
		dbURL string
	)
	flag.StringVar(&dbURL, "db-url", "postgresql:///botsu", "url to connect to postgres db")
	flag.StringVar(&addr, "addr", ":8080", "address to listen on")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite3", "oshistats.db?_journal_mode=WAL")
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

	s := server.NewServer(logRepository, videoVTuberRepository)
	err = http.ListenAndServe(addr, s)
	if err != nil {
		log.Fatalln(err)
	}
}
