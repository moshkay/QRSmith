// Package shortener provides URL shortening with click tracking. Two backends
// implement the Store interface: an in-memory store (default, non-persistent)
// and a MongoDB store (persistent across restarts).
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
	Code      string    `json:"code" bson:"code"`
	URL       string    `json:"url" bson:"url"`
	Clicks    int64     `json:"clicks" bson:"clicks"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// Store abstracts a link backend so the server can use either the in-memory or
// the MongoDB implementation interchangeably.
type Store interface {
	// Create validates rawURL and stores it under a freshly generated code.
	Create(rawURL string) (*Link, error)
	// Resolve returns the link for a code and increments its click count.
	Resolve(code string) (*Link, bool)
	// Stats returns the link for a code without incrementing clicks.
	Stats(code string) (*Link, bool)
}

// MemoryStore is a concurrency-safe in-memory link store. Links are lost on
// restart; use MongoStore for persistence.
type MemoryStore struct {
	mu    sync.RWMutex
	links map[string]*Link
}

// NewMemoryStore creates an empty in-memory link store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: make(map[string]*Link)}
}

// Create validates rawURL and stores it under a freshly generated unique code.
func (s *MemoryStore) Create(rawURL string) (*Link, error) {
	normalized, verr := validateURL(rawURL)
	if verr != nil {
		return nil, verr
	}

	code, err := s.uniqueCode()
	if err != nil {
		return nil, &Error{Code: "INTERNAL_ERROR", Message: "Could not allocate a short code"}
	}

	link := &Link{Code: code, URL: normalized, CreatedAt: time.Now().UTC()}

	s.mu.Lock()
	s.links[code] = link
	s.mu.Unlock()
	return link, nil
}

// Resolve returns the link for a code and atomically increments its click count.
func (s *MemoryStore) Resolve(code string) (*Link, bool) {
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
func (s *MemoryStore) Stats(code string) (*Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	link, ok := s.links[code]
	return link, ok
}

func (s *MemoryStore) uniqueCode() (string, error) {
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

// validateURL trims and validates a raw URL, returning the normalized form.
func validateURL(rawURL string) (string, *Error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", &Error{Code: "MISSING_URL", Message: "A URL is required"}
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", &Error{Code: "INVALID_URL", Message: "Provide a valid http(s) URL"}
	}
	return parsed.String(), nil
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
