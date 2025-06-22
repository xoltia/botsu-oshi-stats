package server

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/xoltia/botsu-oshi-stats/hololist"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server/components"
)

type Server struct {
	vtubers *hololist.Store
	logs    *logs.Source
}

func NewServer(vtubers *hololist.Store, logs *logs.Source) *Server {
	return &Server{vtubers, logs}
}

func keyFromParams(date string, id string) *logs.PaginationKey {
	us, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		return nil
	}
	timePart := time.Unix(0, us*1000)
	idPart, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil
	}
	return &logs.PaginationKey{Date: timePart, ID: idPart}
}

func encodeKeyURL(referring *url.URL, next *logs.PaginationKey) string {
	if next == nil {
		return ""
	}

	newURL := *referring
	us := next.Date.UnixMicro()
	usString := strconv.FormatInt(us, 10)
	idString := strconv.FormatInt(next.ID, 10)

	newQuery := newURL.Query()
	newQuery.Set("ts-offset", usString)
	newQuery.Set("id-offset", idString)
	newURL.RawQuery = newQuery.Encode()

	return newURL.String()
}

func (s *Server) GetIndex(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userID := query.Get("user")
	date := query.Get("ts-offset")
	id := query.Get("id-offset")

	key := keyFromParams(date, id)
	userLogs, err := s.logs.GetByUser(r.Context(), userID, key, 10)
	if err != nil {
		log.Printf("Error getting user logs: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	nextKey := logs.PaginationKeyFromSet(userLogs)
	keyURL := encodeKeyURL(r.URL, nextKey)
	components.HomePage(userLogs, templ.SafeURL(keyURL)).Render(r.Context(), w)
}
