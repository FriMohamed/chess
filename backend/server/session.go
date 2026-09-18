package server

import "github.com/google/uuid"

func (s *Server) createSession(gameID, playerID string) *Session {
	session := &Session{
		ID:       uuid.NewString(),
		GameID:   gameID,
		PlayerID: playerID,
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return session
}

func (s *Server) getSession(sessionID string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessions[sessionID]
}

func (s *Server) deleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}