package session

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"sync"
	"time"
)

type Session struct {
	id           string
	user         *model.User
	project      *model.Project
	createdAt    time.Time
	lastActivity time.Time
	mutex        sync.RWMutex
}

func (s *Session) ID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.id
}

func (s *Session) User() *model.User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.user
}

func (s *Session) Project() *model.Project {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.project
}

func (s *Session) SetProject(project *model.Project) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.project = project
}

func (s *Session) expired(m *Manager) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return time.Since(s.lastActivity) >= m.idleExpiration || time.Since(s.createdAt) >= m.absoluteExpiration
}
