package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/xoltia/botsu-oshi-stats/components"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

type server struct {
	vtubers *vtubers.Store
	logs    *logs.Source
}

func newServer(vtubers *vtubers.Store, logs *logs.Source) *server {
	return &server{vtubers, logs}
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

func encodeKeyURL(path, user string, next *logs.PaginationKey) string {
	if next == nil {
		return ""
	}

	us := next.Date.UnixMicro()
	usString := strconv.FormatInt(us, 10)
	idString := strconv.FormatInt(next.ID, 10)

	newURL := url.URL{Path: path}
	newQuery := url.Values{}
	newQuery.Set("user", user)
	newQuery.Set("ts-offset", usString)
	newQuery.Set("id-offset", idString)
	newURL.RawQuery = newQuery.Encode()

	return newURL.String()
}

func (s *server) getLogContinuation(w http.ResponseWriter, r *http.Request) {
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
	keyURL := encodeKeyURL("/log-continuation", userID, nextKey)

	logDetails := make([]components.LogDetails, len(userLogs))
	for i, userLog := range userLogs {
		vtubers, err := detectVTubers(r.Context(), s.vtubers, userLog)
		if err != nil {
			log.Printf("Error searching vtubers: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logDetails[i] = components.LogDetails{Log: userLog, VTubers: vtubers}

	}

	components.LogsPart(logDetails, templ.SafeURL(keyURL)).Render(r.Context(), w)
}

func (s *server) getIndex(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userID := query.Get("user")

	userLogs, err := s.logs.GetByUser(r.Context(), userID, nil, 10)
	if err != nil {
		log.Printf("Error getting user logs: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	nextKey := logs.PaginationKeyFromSet(userLogs)
	keyURL := encodeKeyURL("/log-continuation", userID, nextKey)
	logDetails := make([]components.LogDetails, len(userLogs))

	for i, userLog := range userLogs {
		vtubers, err := detectVTubers(r.Context(), s.vtubers, userLog)
		if err != nil {
			log.Printf("Error searching vtubers: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logDetails[i] = components.LogDetails{Log: userLog, VTubers: vtubers}
	}

	components.HomePage(logDetails, templ.SafeURL(keyURL)).Render(r.Context(), w)
}

func findVTuberByHandleOrID(ctx context.Context, s *vtubers.Store, idOrHandle string) (vtubers.VTuber, error) {
	var (
		vtuber vtubers.VTuber
		err    error
	)

	if strings.HasPrefix(idOrHandle, "@") {
		vtuber, err = s.FindByYouTubeHandle(ctx, idOrHandle)
	} else {
		vtuber, err = s.FindByYouTubeHandle(ctx, idOrHandle)
	}

	return vtuber, err
}

func detectVTubers(ctx context.Context, s *vtubers.Store, log logs.Log) ([]vtubers.VTuber, error) {
	found := make([]vtubers.VTuber, 0)

	vtuber, err := s.FindByYouTubeID(ctx, log.VideoMeta.ChannelID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	} else if err == nil {
		found = append(found, vtuber)
	}

	for _, idOrHandle := range log.VideoMeta.LinkedChannels {
		vtuber, err := findVTuberByHandleOrID(ctx, s, idOrHandle)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}

		existing := false
		for _, v := range found {
			if v.ID == vtuber.ID {
				existing = true
				break
			}
		}
		if !existing {
			found = append(found, vtuber)
		}
	}

	return found, nil
}
