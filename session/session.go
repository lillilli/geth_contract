package session

import "sync"

// UserSessionStore - represents user session store
type UserSessionStore interface {
	AddWatchedTx(session string, tx string)
	GetWatchedTxs(session string) []string
}

type userSessionStore struct {
	cache map[string][]string
	sync.RWMutex
}

// NewUserSessionStore - returns new user session store instance
func NewUserSessionStore() UserSessionStore {
	return &userSessionStore{
		cache: make(map[string][]string),
	}
}

// AddWatchedTx - adds tx to session watch list
func (s *userSessionStore) AddWatchedTx(session string, tx string) {
	s.Lock()

	_, ok := s.cache[session]
	if !ok {
		s.cache[session] = make([]string, 0)
	}

	s.cache[session] = append(s.cache[session], tx)
	s.Unlock()
}

func (s *userSessionStore) GetWatchedTxs(session string) []string {
	s.RLock()
	defer s.RUnlock()

	return s.cache[session]
}
