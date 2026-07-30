// Package server wires the HTTP API and static frontend for QRForge.
package server

import (
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/dojah/qrforge/internal/beautify"
	"github.com/dojah/qrforge/internal/config"
	"github.com/dojah/qrforge/internal/qr"
	"github.com/dojah/qrforge/internal/shortener"
	"github.com/dojah/qrforge/web"
)

// Server holds dependencies and the configured HTTP handler.
type Server struct {
	qr              *qr.Generator
	beautify        *beautify.Service
	shortener       shortener.Store
	maxContentBytes int
	maxBodyBytes    int64
	handler         http.Handler
}

// New builds a fully configured Server from application config and a link store.
func New(cfg config.Config, store shortener.Store) (*Server, error) {
	s := &Server{
		qr:              qr.NewGenerator(),
		beautify:        beautify.NewService(int(cfg.MaxLogoBytes)),
		shortener:       store,
		maxContentBytes: cfg.MaxContentBytes,
		// Allow room for a base64-encoded logo (~4/3 expansion) plus JSON overhead.
		maxBodyBytes: cfg.MaxLogoBytes*2 + 64*1024,
	}

	staticFS, err := fs.Sub(web.Static, "static")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/qr-presets", s.handlePresets)
	mux.HandleFunc("/api/v1/qr-codes", s.handleCreateQRCode)
	mux.HandleFunc("/api/v1/qr-codes/download", s.handleDownloadQRCode)
	mux.HandleFunc("/api/v1/short-links", s.handleCreateShortLink)
	mux.HandleFunc("/api/v1/short-links/", s.handleShortLinkStats)
	mux.HandleFunc("/s/", s.handleRedirect)

	// Clean-URL pages served from embedded HTML. Client-side JS reads the path
	// (e.g. /qr/wifi) to render the right content.
	page := pageHandler(staticFS)
	mux.HandleFunc("/faq", page("faq.html"))
	mux.HandleFunc("/learn", page("learn.html"))
	mux.HandleFunc("/my-links", page("my-links.html"))
	mux.HandleFunc("/qr/", page("guide.html"))

	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	limiter := newRateLimiter(120, time.Minute)
	s.handler = chain(mux,
		recoverer,
		requestLogger,
		securityHeaders,
		limiter.middleware,
	)

	return s, nil
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler { return s.handler }

// pageHandler returns a factory that serves a named embedded HTML file. Used to
// map clean URLs (/faq, /learn, /qr/{slug}) onto static single-page documents.
func pageHandler(fsys fs.FS) func(name string) http.HandlerFunc {
	return func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f, err := fsys.Open(name)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(data)
		}
	}
}
