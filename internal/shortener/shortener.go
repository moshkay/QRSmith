// Package shortener provides an in-memory URL shortening service with click
// tracking. Links are not persisted across restarts.
package shortener

import (
	"crypto/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const codeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

// Error is a user-facing shortener error with a stable code.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

// Link is a shortened URL record.
type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is a concurrency-safe in-memory link store.
type Store struct {
	mu    sync.RWMutex
	links map[string]*Link
}

// NewStore creates an empty link store.
func NewStore() *Store {
	return &Store{links: make(map[string]*Link)}
}

// Create validates rawURL and stores it under a freshly generated unique code.
func (s *Store) Create(rawURL string) (*Link, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, &Error{Code: "MISSING_URL", Message: "A URL is required"}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, &Error{Code: "INVALID_URL", Message: "Provide a valid http(s) URL"}
	}

	code, err := s.uniqueCode()
	if err != nil {
		return nil, &Error{Code: "INTERNAL_ERROR", Message: "Could not allocate a short code"}
	}

	link := &Link{Code: code, URL: parsed.String(), CreatedAt: time.Now().UTC()}

	s.mu.Lock()
	s.links[code] = link
	s.mu.Unlock()
	return link, nil
}

// Resolve returns the link for a code and atomically increments its click count.
func (s *Store) Resolve(code string) (*Link, bool) {
	s.mu.RLock()
	link, ok := s.links[code]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	atomic.AddInt64(&link.Clicks, 1)
	return link, true
}

// Stats returns the link for a code without incrementing clicks.
func (s *Store) Stats(code string) (*Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	link, ok := s.links[code]
	return link, ok
}

func (s *Store) uniqueCode() (string, error) {
	for attempts := 0; attempts < 10; attempts++ {
		code, err := randomCode()
		if err != nil {
			return "", err
		}
		s.mu.RLock()
		_, exists := s.links[code]
		s.mu.RUnlock()
		if !exists {
			return code, nil
		}
	}
	return "", &Error{Code: "INTERNAL_ERROR", Message: "code space exhausted"}
}

func randomCode() (string, error) {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, codeLength)
	for i, b := range buf {
		out[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(out), nil
}
