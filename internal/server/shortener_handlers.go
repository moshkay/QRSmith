package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/dojah/qrforge/internal/shortener"
)

type createShortLinkRequest struct {
	URL string `json:"url"`
}

type shortLinkResponse struct {
	Code      string `json:"code"`
	ShortURL  string `json:"shortUrl"`
	URL       string `json:"url"`
	Clicks    int64  `json:"clicks"`
	CreatedAt string `json:"createdAt"`
}

// handleCreateShortLink creates a short link for a URL.
func (s *Server) handleCreateShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 8*1024))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Could not read request body")
		return
	}

	var req createShortLinkRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return
	}

	link, err := s.shortener.Create(req.URL)
	if err != nil {
		var se *shortener.Error
		if errors.As(err, &se) {
			status := http.StatusUnprocessableEntity
			if se.Code == "INTERNAL_ERROR" {
				status = http.StatusInternalServerError
			}
			writeError(w, status, se.Code, se.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not create short link")
		return
	}

	writeSuccess(w, http.StatusCreated, s.toShortLinkResponse(r, link))
}

// handleShortLinkStats returns click stats for a short link code.
func (s *Server) handleShortLinkStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/api/v1/short-links/")
	if code == "" || strings.Contains(code, "/") {
		writeError(w, http.StatusBadRequest, "INVALID_CODE", "Invalid short link code")
		return
	}

	link, ok := s.shortener.Stats(code)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Short link not found")
		return
	}
	writeSuccess(w, http.StatusOK, s.toShortLinkResponse(r, link))
}

// handleRedirect resolves a short code and 302-redirects to the target URL,
// incrementing the click counter.
func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/s/")
	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		return
	}

	link, ok := s.shortener.Resolve(code)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, link.URL, http.StatusFound)
}

func (s *Server) toShortLinkResponse(r *http.Request, link *shortener.Link) shortLinkResponse {
	return shortLinkResponse{
		Code:      link.Code,
		ShortURL:  buildShortURL(r, link.Code),
		URL:       link.URL,
		Clicks:    link.Clicks,
		CreatedAt: link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// buildShortURL derives the public short URL from the incoming request, honoring
// a reverse-proxy's X-Forwarded-Proto when present.
func buildShortURL(r *http.Request, code string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host + "/s/" + code
}
