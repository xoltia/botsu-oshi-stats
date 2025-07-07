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
	"github.com/xoltia/botsu-oshi-stats/auth"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

func main() {
	var (
		addr        string
		dbURL       string
		oauthConfig server.OAuthConfig
	)
	flag.StringVar(&dbURL, "db-url", "postgresql:///botsu", "url to connect to postgres db")
	flag.StringVar(&addr, "addr", ":8080", "address to listen on")
	flag.StringVar(&oauthConfig.ClientID, "oauth-client-id", "", "discord oauth client id")
	flag.StringVar(&oauthConfig.ClientSecret, "oauth-client-secret", "", "discord oauth client secret")
	flag.StringVar(&oauthConfig.RedirectURL, "oauth-redirect-url", "http://localhost:8080/auth/callback", "discord oauth redirect url")
	flag.Parse()

	if oauthConfig.ClientSecret == "" {
		oauthConfig.ClientSecret = os.Getenv("BOTSU_WEB_OAUTH_CLIENT_SECRET")
	}

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

	repo, err := index.CreateIndexedVideoRepository(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	vtuberStore, err := vtubers.CreateStore(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	sessionStore, err := auth.CreateSessionStore(ctx, db)
	if err != nil {
		log.Panicln(err)
	}

	s := server.NewServer(logRepository, repo, vtuberStore, sessionStore, oauthConfig)
	err = http.ListenAndServe(addr, s)
	if err != nil {
		log.Fatalln(err)
	}
}
