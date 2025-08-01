package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/xoltia/botsu-oshi-stats/auth"
	"github.com/xoltia/botsu-oshi-stats/index"
	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/server/components"
	"github.com/xoltia/botsu-oshi-stats/server/static"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type ImgproxyConfig struct {
	Host string
	Salt []byte
	Key  []byte
}

type Server struct {
	logRepo        *logs.UserLogRepository
	indexRepo      *index.IndexedVideoRepository
	vtuberRepo     *vtubers.Store
	sessions       *auth.SessionStore
	oauthConfig    OAuthConfig
	imgproxyConfig ImgproxyConfig
}

func NewServer(
	logRepo *logs.UserLogRepository,
	indexedVideoRepo *index.IndexedVideoRepository,
	vtuberStore *vtubers.Store,
	sessionStore *auth.SessionStore,
	oauthConfig OAuthConfig,
	imgproxyConfig ImgproxyConfig,
) *Server {
	return &Server{
		logRepo:        logRepo,
		indexRepo:      indexedVideoRepo,
		vtuberRepo:     vtuberStore,
		sessions:       sessionStore,
		oauthConfig:    oauthConfig,
		imgproxyConfig: imgproxyConfig,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHandler := auth.NewDiscordOAuthHandler(s.sessions, auth.DiscordOAuthHandlerConfig{
		ClientID:     s.oauthConfig.ClientID,
		ClientSecret: s.oauthConfig.ClientSecret,
		RedirectURL:  s.oauthConfig.RedirectURL,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", authHandler.WrapHandlerFunc(s.getIndex))
	mux.HandleFunc("GET /logs", authHandler.WrapHandlerFunc(s.getLogs))
	mux.HandleFunc("GET /overview", authHandler.WrapHandlerFunc(s.getOverview))
	mux.HandleFunc("GET /auth/callback", authHandler.HandleCallback)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))
	mux.ServeHTTP(w, r)
}

func (s *Server) getOverview(w http.ResponseWriter, r *http.Request) {
	timelineType := r.URL.Query().Get("type")
	if timelineType != "week" && timelineType != "all" {
		timelineType = "all"
	}
	session := auth.MustSessionFromContext(r.Context())
	model := components.TimelinePageModel{
		Type:                  timelineType,
		UserProfilePictureURL: avatarURL(session),
	}

	var (
		start time.Time
		end   time.Time = time.Now()
	)
	if timelineType == "week" {
		start = end.AddDate(0, 0, -6)
	}

	var (
		history []index.WatchTime
		err     error
	)

	// TODO: based on actual time gap
	if timelineType == "week" {
		history, err = s.indexRepo.GetDailyWatchTimeInRange(r.Context(), session.UserID, start, end)
	} else {
		history, err = s.indexRepo.GetMonthlyWatchTimeInRange(r.Context(), session.UserID, start, end)
	}
	if err != nil {
		log.Printf("Error getting watch time: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, h := range history {
		model.Timeline.Labels = append(model.Timeline.Labels, h.GroupedDate)
		model.Timeline.Values = append(model.Timeline.Values, int(h.Duration.Minutes()))
	}

	topVTubers, err := s.indexRepo.GetTopVTubersByAppearenceCount(r.Context(), session.UserID, start, end, 10)
	if err != nil {
		log.Printf("Error getting top vtubers: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	for _, v := range topVTubers {
		avatarURL := v.PictureURL
		channel, err := s.vtuberRepo.FindChannelByID(r.Context(), v.YouTubeID)
		if err == nil {
			avatarURL = channel.AvatarURL
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("get vtuber channel: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		model.TopVTubersAppearances = append(model.TopVTubersAppearances, components.TopVTuberWithAppearances{
			TopVTuber: components.TopVTuber{
				AvatarURL:    s.getImgproxyURL(avatarURL, "format:webp"),
				Name:         v.EnglishName,
				OriginalName: v.OriginalName,
			},
			Appearances: v.Appearances,
		})
	}
	topVTubersDuration, err := s.indexRepo.GetTopVTubersByDuration(r.Context(), session.UserID, start, end, 10)
	if err != nil {
		log.Printf("Error getting top vtubers by duration: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	for _, v := range topVTubersDuration {
		avatarURL := v.PictureURL
		channel, err := s.vtuberRepo.FindChannelByID(r.Context(), v.YouTubeID)
		if err == nil {
			avatarURL = channel.AvatarURL
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("get vtuber channel: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		model.TopVTubersDuration = append(model.TopVTubersDuration, components.TopVTuberWithDuration{
			TopVTuber: components.TopVTuber{
				AvatarURL:    s.getImgproxyURL(avatarURL, "format:webp"),
				Name:         v.EnglishName,
				OriginalName: v.OriginalName,
			},
			Duration: v.Duration,
		})
	}

	components.TimelinePage(model).Render(r.Context(), w)
}

func (s *Server) getIndex(w http.ResponseWriter, r *http.Request) {
	session := auth.MustSessionFromContext(r.Context())
	userID := session.UserID

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
		video.ThumbnailURL = s.getImgproxyURL(vid.ThumbnailURL, "format:webp", "width:500")
		video.URL = fmt.Sprintf("https://youtu.be/%s", vid.ID)
		video.VTubers = make([]components.WatchedVideoVTuber, len(vtubers))

		for i, vtuber := range vtubers {
			video.VTubers[i] = components.WatchedVideoVTuber{
				OshiMark: vtuber.OshiMark,
				Name:     vtuber.EnglishName,
			}
		}

		videos = append(videos, video)
	}
	continuationURL := getContinuationURL(nextKey)

	// TODO: implement better ranking
	const topVTubersNumber = 6
	topVTubersModel := make([]components.TopVTuber, 0, topVTubersNumber)
	topVTubers, err := s.indexRepo.GetTopVTubersByAppearenceCount(
		r.Context(),
		userID,
		time.Time{},
		time.Now(),
		topVTubersNumber,
	)
	if err != nil {
		log.Printf("get top vtubers: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, v := range topVTubers {
		avatarURL := v.PictureURL
		channel, err := s.vtuberRepo.FindChannelByID(r.Context(), v.YouTubeID)
		if err == nil {
			avatarURL = channel.AvatarURL
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("get vtuber channel: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		topVTubersModel = append(topVTubersModel, components.TopVTuber{
			AvatarURL:    s.getImgproxyURL(avatarURL, "format:webp"),
			Name:         v.EnglishName,
			OriginalName: v.OriginalName,
		})
	}

	end := time.Now()
	start := end.AddDate(0, 0, -6)
	topVTubersModelWeek := make([]components.TopVTuber, 0, topVTubersNumber)
	topVTubersWeek, err := s.indexRepo.GetTopVTubersByAppearenceCount(
		r.Context(),
		userID,
		start,
		end,
		topVTubersNumber,
	)
	if err != nil {
		log.Printf("get top vtubers: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, v := range topVTubersWeek {
		avatarURL := v.PictureURL
		channel, err := s.vtuberRepo.FindChannelByID(r.Context(), v.YouTubeID)
		if err == nil {
			avatarURL = channel.AvatarURL
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("get vtuber channel: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		topVTubersModelWeek = append(topVTubersModelWeek, components.TopVTuber{
			AvatarURL:    s.getImgproxyURL(avatarURL, "format:webp"),
			Name:         v.EnglishName,
			OriginalName: v.OriginalName,
		})
	}

	model := components.IndexPageModel{
		Videos:            videos,
		ContinuationURL:   continuationURL,
		TopVTubersAllTime: topVTubersModel,
		TopVTubersWeekly:  topVTubersModelWeek,
	}

	model.UserProfilePictureURL = avatarURL(session)
	components.IndexPage(model).Render(r.Context(), w)
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustSessionFromContext(r.Context()).UserID
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
		video.ThumbnailURL = s.getImgproxyURL(vid.ThumbnailURL, "format:webp", "width:500")
		video.URL = fmt.Sprintf("https://youtu.be/%s", vid.ID)
		video.VTubers = make([]components.WatchedVideoVTuber, len(vtubers))

		for i, vtuber := range vtubers {
			video.VTubers[i] = components.WatchedVideoVTuber{
				OshiMark: vtuber.OshiMark,
				Name:     vtuber.EnglishName,
			}
		}

		videos = append(videos, video)
	}

	continuationURL := getContinuationURL(nextKey)
	components.WatchedVideoGridElements(videos, continuationURL).Render(r.Context(), w)
}

func avatarURL(session auth.Session) string {
	if session.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.webp?size=128", session.UserID, session.Avatar)
}

func getContinuationURL(key logs.PaginationKey) string {
	if key.IsZero() {
		return ""
	}
	u := url.URL{Path: "/logs"}
	q := url.Values{}
	q.Set("cursor", key.EncodeBase64String())
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Server) getImgproxyURL(originalURL string, options ...string) string {
	if s.imgproxyConfig.Host == "" {
		return originalURL
	}

	optionsLen := 0
	for _, o := range options {
		optionsLen += len(o)
	}

	path := []byte{}
	for _, o := range options {
		path = append(path, '/')
		path = append(path, o...)
	}
	path = append(path, '/')
	path = base64.RawURLEncoding.AppendEncode(path, []byte(originalURL))

	mac := hmac.New(sha256.New, s.imgproxyConfig.Key)
	mac.Write(s.imgproxyConfig.Salt)
	mac.Write(path)
	signature := mac.Sum(nil)

	return fmt.Sprintf(
		"https://%s/%s%s",
		s.imgproxyConfig.Host,
		base64.RawURLEncoding.EncodeToString(signature),
		path)
}
