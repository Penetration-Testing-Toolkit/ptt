package session

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	id           string
	userID       uuid.UUID
	username     string
	isAdmin      bool
	createdAt    time.Time
	lastActivity time.Time
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) UserID() uuid.UUID {
	return s.userID
}

func (s *Session) Username() string {
	return s.username
}

func (s *Session) IsAdmin() bool {
	return s.isAdmin
}

func (s *Session) expired(m *Manager) bool {
	return time.Since(s.lastActivity) >= m.idleExpiration || time.Since(s.createdAt) >= m.absoluteExpiration
}
