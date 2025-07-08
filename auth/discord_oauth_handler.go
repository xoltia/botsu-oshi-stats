package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type DiscordOAuthHandlerConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type DiscordOAuthHandler struct {
	sessions *SessionStore
	config   *oauth2.Config
}

func NewDiscordOAuthHandler(
	sessions *SessionStore,
	config DiscordOAuthHandlerConfig,
) *DiscordOAuthHandler {
	return &DiscordOAuthHandler{
		sessions: sessions,
		config: &oauth2.Config{
			RedirectURL:  config.RedirectURL,
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://discord.com/api/oauth2/authorize",
				TokenURL:  "https://discord.com/api/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{"identify"},
		},
	}
}

func (h *DiscordOAuthHandler) handleNewSession(w http.ResponseWriter, r *http.Request) {
	session, err := generateRandomSession()
	if err != nil {
		log.Printf("Error generating session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.sessions.insertSession(r.Context(), session); err != nil {
		log.Printf("Error inserting session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     "sessionID",
		Value:    session.ID,
	})

	url := h.config.AuthCodeURL(session.OAuthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *DiscordOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("sessionID")
	if err != nil {
		h.handleNewSession(w, r)
		return
	}

	sessionID := sessionCookie.Value
	session, err := h.sessions.findSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.handleNewSession(w, r)
			return
		}
		log.Printf("Error finding session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.FormValue("state") != session.OAuthState {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State does not match."))
		return
	}

	token, err := h.config.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error getting token: %s", err)
		return
	}

	res, err := h.config.
		Client(r.Context(), token).
		Get("https://discord.com/api/users/@me")
	if err != nil || res.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error getting user: %s", err)
		return
	}
	defer res.Body.Close()

	var user struct {
		ID     string `json:"id"`
		Avatar string `json:"avatar"`
	}

	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error reading user fetch response: %s", err)
		return
	}

	err = h.sessions.setSessionUserData(r.Context(), sessionID, user.ID, user.Avatar)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error setting session user: %s", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *DiscordOAuthHandler) WrapHandlerFunc(inner http.HandlerFunc) http.HandlerFunc {
	return h.WrapHandler(inner).ServeHTTP
}

// WrapHandler wraps an http handler, ensuring that only the inner handler
// is used for requests with an authenticated session.
func (h *DiscordOAuthHandler) WrapHandler(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("sessionID")
		if err != nil {
			h.handleNewSession(w, r)
			return
		}

		sessionID := sessionCookie.Value
		session, err := h.sessions.findSession(r.Context(), sessionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				h.handleNewSession(w, r)
				return
			}
			log.Printf("Error finding session: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if session.UserID == "" {
			url := h.config.AuthCodeURL(session.OAuthState)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		ctx := contextWithSession(r.Context(), session)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
