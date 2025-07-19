package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
	"golang.org/x/time/rate"
)

func main() {
	options := vtubers.UpdateOptions{}
	flag.StringVar(&options.GoogleAPIKey, "google-api-key", "", "google api key for youtube data api")
	flag.BoolVar(&options.ChannelsOnly, "channels-only", false, "only update channel data")
	flag.Parse()

	if options.GoogleAPIKey == "" {
		options.GoogleAPIKey = os.Getenv("BOTSU_WEB_GOOGLE_API_KEY")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite3", "oshistats.db?_journal_mode=WAL")
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	client := &http.Client{}
	limiter := rate.NewLimiter(rate.Limit(time.Second), 2)
	scraper := vtubers.NewHololistScraper(client, limiter)
	vtuberStore, err := vtubers.CreateStore(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	updater := vtubers.Updater{
		Scraper: scraper,
		Store:   vtuberStore,
		Options: options,
	}

	err = updater.Update(ctx)
	if err != nil {
		log.Panicln(err)
	}
}
