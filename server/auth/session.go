package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
)

type Session struct {
	ID        string
	UserID    int
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) CreateSession(user *domain.User, duration time.Duration) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = session
	return session, nil
}

func (s *SessionStore) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, sessionID)
		return nil, false
	}

	return session, true
}

func (s *SessionStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

func (s *SessionStore) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, sessionID)
		}
	}
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
