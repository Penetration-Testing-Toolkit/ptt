package session

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
	"time"
)

// Manager based loosely on https://themsaid.com/building-secure-session-manager-in-go
type Manager struct {
	store              map[string]*Session
	idleExpiration     time.Duration
	absoluteExpiration time.Duration
	cookieName         string
	logger             hclog.Logger
}

// gc cleans up expired sessions on a time.Ticker.
func (m *Manager) gc(d time.Duration) {
	ticker := time.NewTicker(d)

	for range ticker.C {
		for k, s := range m.store {
			if s.expired(m) {
				// Delete key and associated value
				delete(m.store, k)
			}
		}
	}
}

// valid checks if a sessionID exists and is not expired.
// If a session is found but is expired, it is deleted.
func (m *Manager) valid(sessionID string) bool {
	s, ok := m.store[sessionID]
	if !ok {
		return false
	}

	if s.expired(m) {
		// Delete expired key and associated value
		delete(m.store, sessionID)
		return false
	}

	return true
}

// setCookie creates the session cookie on the response.
func (m *Manager) setCookie(w http.ResponseWriter, id string, expiration time.Duration) {
	w.Header().Add("Vary", "Cookie")
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)

	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    id,
		Path:     "/",
		Expires:  time.Now().Add(expiration),
		MaxAge:   int(expiration / time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}

func (m *Manager) invalidateCookie(w http.ResponseWriter, id string) {
	m.setCookie(w, id, time.Duration(0))
}

// getIDFromCookie attempts to get the session ID from the request's cookies.
func (m *Manager) getIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// findSession returns the Session with the given sessionID.
// If the Session does not exist or is invalid the returned Session will be nil.
func (m *Manager) findSession(sessionID string) *Session {
	valid := m.valid(sessionID)
	if !valid {
		return nil
	}

	return m.store[sessionID]
}

func (m *Manager) NewSession(w http.ResponseWriter, userID uuid.UUID) {
	id := randomID()
	i := 0
	for {
		i++
		if i >= 100 {
			panic("stuck! unable to generate an unused session ID")
		}

		if _, ok := m.store[id]; ok {
			// rare collision, generate a new ID
			id = randomID()
		} else {
			break
		}
	}

	s := &Session{
		id:           id,
		userID:       userID,
		createdAt:    time.Now(),
		lastActivity: time.Now(),
	}

	m.store[id] = s

	m.setCookie(w, id, m.idleExpiration)

	return
}

func (m *Manager) DeleteSession(sessionID string) {
	delete(m.store, sessionID)
}

func (m *Manager) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip this middleware on static assets and the login page
		if strings.HasPrefix(c.Request().URL.Path, "/static") ||
			c.Request().URL.Path == "/favicon.ico" ||
			c.Request().URL.Path == "/login" {
			return next(c)
		}

		// Look for session in cookies
		sessionID := m.getIDFromCookie(c.Request())
		if sessionID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		// Look for session in store
		s := m.findSession(sessionID)
		if s == nil {
			// Invalidate provided session and redirect
			m.invalidateCookie(c.Response(), sessionID)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		// Update session's last use & client's cookie
		s.lastActivity = time.Now()
		m.setCookie(c.Response(), sessionID, m.idleExpiration)

		// Add session to context for use by handlers
		c.Set("session", s)

		return next(c)
	}
}

func NewSessionManager(gcInterval, idleExpiration, absoluteExpiration time.Duration, cookieName string, logger hclog.Logger) *Manager {
	m := &Manager{
		store:              make(map[string]*Session),
		idleExpiration:     idleExpiration,
		absoluteExpiration: absoluteExpiration,
		cookieName:         cookieName,
		logger:             logger,
	}

	// Garbage collect in a new goroutine
	go m.gc(gcInterval)

	return m
}

func randomID() string {
	// 256 bits of random data
	id := make([]byte, 32)

	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic("could not generate random Base64")
	}

	return base64.RawURLEncoding.EncodeToString(id)
}
