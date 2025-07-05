package server

import (
	"net/http"

	"github.com/xoltia/botsu-oshi-stats/server/components"
	"github.com/xoltia/botsu-oshi-stats/server/static"
)

type Server struct{}

func (s Server) getIndex(w http.ResponseWriter, r *http.Request) {
	components.IndexPage().Render(r.Context(), w)
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.getIndex)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))
	mux.ServeHTTP(w, r)
}
