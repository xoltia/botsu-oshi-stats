package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/xoltia/botsu-oshi-stats/hololist"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server"
	"golang.org/x/time/rate"
	_ "modernc.org/sqlite"
)

var (
	updateInterval = flag.Duration("update-interval", time.Hour*48, "how often to check for hololist updates")
)

func updateVTubers(ctx context.Context, updater *hololist.Updater) {
	for {
		lastModified, err := updater.LastUpdate(ctx)
		if err != nil {
			log.Printf("Unable to check last modified time: %s\n", lastModified.String())
			break
		}

		remaining := *updateInterval - time.Since(lastModified)
		if remaining > 0 {
			log.Printf("Updated recently, waiting: %s\n", remaining.String())
			select {
			case <-time.After(remaining):
			case <-ctx.Done():
				return
			}
		}

		err = updater.Update(ctx)
		if err != nil {
			// Keep retrying every 10 minutes if updating fails
			log.Printf("Failed to update: %s\n", err)
			select {
			case <-time.After(time.Minute * 10):
			case <-ctx.Done():
				return
			}
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Setup postgres database
	remoteDB, err := sqlx.Open("pgx", os.Getenv("BOTSU_DB_URL"))
	if err != nil {
		log.Panicln(err)
	}
	defer remoteDB.Close()

	logStore := logs.NewSource(remoteDB)

	// Setup sqlite database and scraper
	client := &http.Client{}
	limiter := rate.NewLimiter(rate.Every(time.Second), 2)
	scraper := hololist.NewScraper(client, limiter)
	dbx, err := sqlx.Open("sqlite", "oshistats.db")
	if err != nil {
		log.Panicln(err)
	}
	defer dbx.Close()

	vtuberStore, err := hololist.CreateStore(ctx, dbx)
	if err != nil {
		log.Panicln(err)
	}

	updater, err := hololist.OpenUpdater(ctx, dbx, vtuberStore, scraper)
	if err != nil {
		log.Panicln(err)
	}

	go updateVTubers(ctx, updater)

	// Setup server
	server := server.NewServer(vtuberStore, logStore)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", server.GetIndex)

	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Panicln(err)
		}
	}()

	<-ctx.Done()
}
