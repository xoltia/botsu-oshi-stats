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
	"github.com/xoltia/botsu-oshi-stats/internal/vtubers"
	"golang.org/x/time/rate"
	_ "modernc.org/sqlite"
)

func main() {
	options := vtubers.UpdateOptions{}
	flag.StringVar(&options.GoogleAPIKey, "google-api-key", "", "google api key for youtube data api")
	flag.BoolVar(&options.ChannelsOnly, "channels-only", false, "only update channel data")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite", "oshistats.db")
	if err != nil {
		log.Panicln(err)
	}

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
