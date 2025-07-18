package session

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Manager based loosely on https://themsaid.com/building-secure-session-manager-in-go
type Manager struct {
	store              sync.Map
	idleExpiration     time.Duration
	absoluteExpiration time.Duration
	cookieName         string
	logger             hclog.Logger
}

// gc cleans up expired sessions on a time.Ticker.
func (m *Manager) gc(d time.Duration) {
	ticker := time.NewTicker(d)

	for range ticker.C {
		m.logger.Debug("session GC starting...")
		m.store.Range(func(k, v any) bool {
			key, ok := k.(string)
			if !ok {
				m.logger.Error("session gc: key is not a string")
				return true
			}

			sess, ok := v.(*Session)
			if !ok {
				m.logger.Error("session gc: value is not a session")
				return true
			}

			if sess.expired(m) {
				// Delete key and associated value
				m.logger.Debug("deleting session", "key", k, "session ID", sess.ID(),
					"username", sess.User().Username)
				m.store.Delete(key)
			}

			return true
		})
		m.logger.Debug("session GC ended")
	}
}

// getValidSession returns a Session if the sessionID exists and is not expired.
// If a Session is found but is expired, it is deleted.
func (m *Manager) getValidSession(sessionID string) *Session {
	s, ok := m.store.Load(sessionID)
	if !ok {
		return nil
	}

	sess, ok := s.(*Session)
	if !ok {
		m.logger.Error("session getValidSession: value is not a session")
		return nil
	}

	if sess.expired(m) {
		// Delete expired key and associated value
		m.store.Delete(sessionID)
		return nil
	}

	return sess
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

// invalidateCookie sets the session cookie's duration to 0 to invalidate it.
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

func (m *Manager) NewSession(w http.ResponseWriter, user *model.User) *Session {
	id := randomID()
	i := 0
	for {
		i++
		if i >= 100 {
			panic("stuck! unable to generate an unused session ID")
		}

		if _, ok := m.store.Load(id); ok {
			// rare collision, generate a new ID
			id = randomID()
		} else {
			break
		}
	}

	s := &Session{
		id:           id,
		user:         user,
		project:      nil,
		createdAt:    time.Now(),
		lastActivity: time.Now(),
		mutex:        sync.RWMutex{},
	}

	m.store.Store(id, s)

	m.setCookie(w, id, m.idleExpiration)

	return s
}

// DeleteSession deletes the Session with matching sessionID from the Manager & invalidates the user's session cookie.
func (m *Manager) DeleteSession(c echo.Context, sessionID string) {
	m.invalidateCookie(c.Response(), sessionID)
	m.store.Delete(sessionID)
}

func (m *Manager) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip this middleware on static assets and the login page
		p := c.Request().URL.Path
		if strings.HasPrefix(p, "/static") ||
			p == "/favicon.ico" ||
			p == "/login" {
			return next(c)
		}

		// Look for session in cookies
		sessionID := m.getIDFromCookie(c.Request())
		if sessionID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		// Look for session in store
		s := m.getValidSession(sessionID)
		if s == nil {
			// Invalidate provided session cookie and send error page
			m.invalidateCookie(c.Response(), sessionID)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		// Update session's last use & client's cookie
		s.mutex.Lock()
		s.lastActivity = time.Now()
		s.mutex.Unlock()
		m.setCookie(c.Response(), sessionID, m.idleExpiration)

		// Add session to context for use by handlers
		c.Set("session", s)

		return next(c)
	}
}

func NewSessionManager(gcInterval, idleExpiration, absoluteExpiration time.Duration, cookieName string, logger hclog.Logger) *Manager {
	m := &Manager{
		store:              sync.Map{},
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
