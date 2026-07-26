package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dojah/qrforge/internal/beautify"
	"github.com/dojah/qrforge/internal/qrcontent"
)

// generateRequest is the request body for creating a QR code. It embeds the
// style request so all styling fields live at the top level of the JSON.
//
// Content resolution order:
//  1. If Content is set, it is used verbatim.
//  2. Otherwise Type + Data are formatted by the qrcontent service.
type generateRequest struct {
	Content string            `json:"content"`
	Type    string            `json:"type"`
	Data    map[string]string `json:"data"`
	Format  string            `json:"format"` // "png" (default) or "svg"
	beautify.StyleRequest
}

// generateResponse is returned by the JSON generation endpoint. The image is a
// data URL so browsers can render it directly in an <img> tag.
type generateResponse struct {
	Image  string `json:"image"`  // data URI
	Format string `json:"format"` // "png" or "svg"
	Width  int    `json:"width"`  // pixels
}

// renderResult carries the encoded image plus metadata for the two endpoints.
type renderResult struct {
	data        []byte
	contentType string
	format      string
	ext         string
	width       int
}

// handleHealth is a liveness/readiness probe.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"service": "qrforge",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handlePresets lists the built-in beautify presets.
func (s *Server) handlePresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	writeSuccess(w, http.StatusOK, beautify.Presets())
}

// handleCreateQRCode generates a QR code and returns it as a base64 data URL.
func (s *Server) handleCreateQRCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	res, ok := s.renderFromRequest(w, r)
	if !ok {
		return
	}

	dataURL := "data:" + res.contentType + ";base64," + base64.StdEncoding.EncodeToString(res.data)
	writeSuccess(w, http.StatusCreated, generateResponse{
		Image:  dataURL,
		Format: res.format,
		Width:  res.width,
	})
}

// handleDownloadQRCode generates a QR code and returns it as a file attachment.
func (s *Server) handleDownloadQRCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	res, ok := s.renderFromRequest(w, r)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", res.contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="qrcode.%s"`, res.ext))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(res.data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res.data)
}

// renderFromRequest parses, validates, and renders a QR code from the request
// body. On any failure it writes the error response and returns ok=false.
func (s *Server) renderFromRequest(w http.ResponseWriter, r *http.Request) (renderResult, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, s.maxBodyBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Could not read request body")
		return renderResult{}, false
	}

	var req generateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON")
		return renderResult{}, false
	}

	content, ok := s.resolveContent(w, req)
	if !ok {
		return renderResult{}, false
	}

	style, err := s.beautify.Resolve(req.StyleRequest)
	if err != nil {
		var ve *beautify.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusUnprocessableEntity, ve.Code, ve.Message)
			return renderResult{}, false
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not process style")
		return renderResult{}, false
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "svg" {
		svg, err := s.qr.RenderSVG(content, style)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "RENDER_FAILED", "Could not generate a QR code for the given input")
			return renderResult{}, false
		}
		return renderResult{data: svg, contentType: "image/svg+xml", format: "svg", ext: "svg", width: style.Size}, true
	}

	png, err := s.qr.Render(content, style)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "RENDER_FAILED", "Could not generate a QR code for the given input")
		return renderResult{}, false
	}
	return renderResult{data: png, contentType: "image/png", format: "png", ext: "png", width: style.Size}, true
}

// resolveContent returns the QR payload string, either verbatim from Content or
// built from Type + Data by the content service.
func (s *Server) resolveContent(w http.ResponseWriter, req generateRequest) (string, bool) {
	content := strings.TrimSpace(req.Content)

	if content == "" && req.Type != "" {
		built, err := qrcontent.Build(req.Type, req.Data)
		if err != nil {
			var be *qrcontent.BuildError
			if errors.As(err, &be) {
				writeError(w, http.StatusUnprocessableEntity, be.Code, be.Message)
				return "", false
			}
			writeError(w, http.StatusUnprocessableEntity, "INVALID_CONTENT", "Could not build QR content")
			return "", false
		}
		content = built
	}

	if content == "" {
		writeError(w, http.StatusUnprocessableEntity, "MISSING_CONTENT", "Content is required")
		return "", false
	}
	if len(content) > s.maxContentBytes {
		writeError(w, http.StatusUnprocessableEntity, "CONTENT_TOO_LONG", "Content exceeds the maximum allowed length")
		return "", false
	}
	return content, true
}
