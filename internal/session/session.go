package session

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"time"
)

type Session struct {
	id           string
	user         *model.User
	project      *model.Project
	createdAt    time.Time
	lastActivity time.Time
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) User() *model.User {
	return s.user
}

func (s *Session) Project() *model.Project {
	return s.project
}

func (s *Session) SetProject(project *model.Project) {
	s.project = project
}

func (s *Session) expired(m *Manager) bool {
	return time.Since(s.lastActivity) >= m.idleExpiration || time.Since(s.createdAt) >= m.absoluteExpiration
}
