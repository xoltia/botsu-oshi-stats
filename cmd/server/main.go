package main

import (
	"context"
	"encoding/hex"
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
		addr                string
		dbURL               string
		oauthConfig         server.OAuthConfig
		imgproxyConfig      server.ImgproxyConfig
		imgproxySigningKey  string
		imgproxySigningSalt string
	)
	flag.StringVar(&dbURL, "db-url", "postgresql:///botsu", "url to connect to postgres db")
	flag.StringVar(&addr, "addr", ":8080", "address to listen on")
	flag.StringVar(&oauthConfig.ClientID, "oauth-client-id", "", "discord oauth client id")
	flag.StringVar(&oauthConfig.ClientSecret, "oauth-client-secret", "", "discord oauth client secret")
	flag.StringVar(&oauthConfig.RedirectURL, "oauth-redirect-url", "http://localhost:8080/auth/callback", "discord oauth redirect url")
	flag.StringVar(&imgproxyConfig.Host, "imgproxy-host", "", "imgproxy host")
	flag.StringVar(&imgproxySigningKey, "imgproxy-key", "", "imgproxy signing key")
	flag.StringVar(&imgproxySigningSalt, "imgproxy-salt", "", "imgproxy signing salt")
	flag.Parse()

	var err error
	if imgproxyConfig.Key, err = hex.DecodeString(imgproxySigningKey); err != nil {
		log.Printf("Invalid imgproxy key value: %s", err)
		return
	}

	if imgproxyConfig.Salt, err = hex.DecodeString(imgproxySigningSalt); err != nil {
		log.Printf("Invalid imgproxy salt value: %s", err)
		return
	}

	if oauthConfig.ClientSecret == "" {
		oauthConfig.ClientSecret = os.Getenv("BOTSU_WEB_OAUTH_CLIENT_SECRET")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := sqlx.Open("sqlite3", "oshistats.db?_journal_mode=WAL")
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

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

	indexedVideoRepo, err := index.CreateIndexedVideoRepository(ctx, db)
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

	s := server.NewServer(
		logRepository,
		indexedVideoRepo,
		vtuberStore,
		sessionStore,
		oauthConfig,
		imgproxyConfig,
	)
	err = http.ListenAndServe(addr, s)
	if err != nil {
		log.Fatalln(err)
	}
}
