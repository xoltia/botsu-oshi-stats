package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server/components"
	"github.com/xoltia/botsu-oshi-stats/server/static"
)

type Server struct {
	logRepo   *logs.UserLogRepository
	indexRepo *index.IndexedVideoRepository
}

func NewServer(lr *logs.UserLogRepository, vr *index.IndexedVideoRepository) *Server {
	return &Server{lr, vr}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.getIndex)
	mux.HandleFunc("GET /logs", s.getLogs)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))
	mux.ServeHTTP(w, r)
}

func (s *Server) getIndex(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	userLogs, nextKey, err := s.logRepo.GetRecentUserVideos(r.Context(), logs.GetRecentUserVideosParams{
		UserID: userID,
		Limit:  12,
	})
	if err != nil {
		log.Printf("get recent user videos error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	videos := make([]components.WatchedVideo, 0, 12)
	for _, vid := range userLogs {
		vtubers, err := s.indexRepo.GetVTubersForVideo(r.Context(), userID, vid.ID)
		if err != nil {
			log.Printf("get video vtubers error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		watchTime, err := s.logRepo.GetTotalVideoWatchTime(r.Context(), userID, vid.ID)
		if err != nil {
			log.Printf("get video watch time error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var video components.WatchedVideo
		video.Title = vid.Title
		video.ChannelTitle = vid.ChannelName
		video.PercentWatched = min(1, float64(watchTime)/float64(vid.Duration))
		video.ThumbnailURL = templ.SafeURL(vid.ThumbnailURL) // TODO: validate url
		video.URL = templ.SafeURL(fmt.Sprintf("https://youtu.be/%s", vid.ID))
		video.VTubers = make([]components.WatchedVideoVTuber, len(vtubers))

		for i, vtuber := range vtubers {
			video.VTubers[i] = components.WatchedVideoVTuber{
				OshiMark: vtuber.OshiMark,
				Name:     vtuber.EnglishName,
			}
		}

		videos = append(videos, video)
	}

	continuationURL := getContinuationURL(userID, nextKey)
	components.IndexPage(videos, templ.SafeURL(continuationURL)).Render(r.Context(), w)
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	cursor := r.URL.Query().Get("cursor")

	var key logs.PaginationKey
	err := key.DecodeBase64String(cursor)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userLogs, nextKey, err := s.logRepo.GetRecentUserVideos(r.Context(), logs.GetRecentUserVideosParams{
		UserID:  userID,
		Limit:   12,
		PageKey: key,
	})

	if err != nil {
		log.Printf("get recent user videos error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	videos := make([]components.WatchedVideo, 0, 12)
	for _, vid := range userLogs {
		vtubers, err := s.indexRepo.GetVTubersForVideo(r.Context(), userID, vid.ID)
		if err != nil {
			log.Printf("get video vtubers error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		watchTime, err := s.logRepo.GetTotalVideoWatchTime(r.Context(), userID, vid.ID)
		if err != nil {
			log.Printf("get video watch time error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var video components.WatchedVideo
		video.Title = vid.Title
		video.ChannelTitle = vid.ChannelName
		video.PercentWatched = min(1, float64(watchTime)/float64(vid.Duration))
		video.ThumbnailURL = templ.SafeURL(vid.ThumbnailURL) // TODO: validate url
		video.URL = templ.SafeURL(fmt.Sprintf("https://youtu.be/%s", vid.ID))
		video.VTubers = make([]components.WatchedVideoVTuber, len(vtubers))

		for i, vtuber := range vtubers {
			video.VTubers[i] = components.WatchedVideoVTuber{
				OshiMark: vtuber.OshiMark,
				Name:     vtuber.EnglishName,
			}
		}

		videos = append(videos, video)
	}

	continuationURL := getContinuationURL(userID, nextKey)
	components.WatchedVideoGridElements(videos, templ.SafeURL(continuationURL)).Render(r.Context(), w)
}

func getContinuationURL(userID string, key logs.PaginationKey) string {
	if key.IsZero() {
		return ""
	}
	u := url.URL{Path: "/logs"}
	q := url.Values{}
	q.Set("user", userID)
	q.Set("cursor", key.EncodeBase64String())
	u.RawQuery = q.Encode()
	return u.String()
}
