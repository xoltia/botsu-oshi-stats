package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/xoltia/botsu-oshi-stats/internal/logs"
	"github.com/xoltia/botsu-oshi-stats/internal/vtubers"
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

	detecter, err := vtubers.CreateDetector(ctx, store)
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

	i := 0
	for ls.Next() && i < 10 {
		l, err := ls.Scan()
		if err != nil {
			log.Panicln(err)
		}

		vtubers, err := detecter.Detect(ctx, l)
		if err != nil {
			log.Panicln(err)
		}

		fmt.Printf("https://youtube.com/watch?v=%v\n", l.Video.ID)
		fmt.Printf("%v\n", vtubers.All)
		i++
	}

	if err := ls.Err(); err != nil {
		log.Panicln(err)
	}
}
