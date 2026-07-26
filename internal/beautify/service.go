// Package beautify is the styling service. It validates and normalizes
// user-supplied style requests, applies preset themes, decodes logos, and
// guards scannability (e.g. contrast) before handing a clean qr.Style to the
// rendering engine.
package beautify

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"math"
	"regexp"
	"strings"

	"github.com/dojah/qrforge/internal/qr"
)

const (
	minSize     = 128
	maxSize     = 2048
	defaultSize = 512

	maxBorder     = 256
	defaultBorder = 30

	// minContrastRatio is the lowest fg/bg luminance contrast we allow so the
	// generated code stays reliably scannable (WCAG-style ratio, 1.0 .. 21.0).
	minContrastRatio = 1.6
)

var hexColorPattern = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// GradientRequest is the wire form of a foreground gradient.
type GradientRequest struct {
	Angle float64           `json:"angle"`
	Stops []GradientStopReq `json:"stops"`
}

// GradientStopReq is a single gradient color stop from the client.
type GradientStopReq struct {
	Offset float64 `json:"offset"`
	Color  string  `json:"color"`
}

// StyleRequest is the raw, untrusted style payload from an API consumer.
type StyleRequest struct {
	Preset          string           `json:"preset"`
	Foreground      string           `json:"foreground"`
	Background      string           `json:"background"`
	Transparent     bool             `json:"transparent"`
	Shape           string           `json:"shape"`
	ErrorCorrection string           `json:"errorCorrection"`
	Size            int              `json:"size"`
	BorderWidth     *int             `json:"borderWidth"`
	Gradient        *GradientRequest `json:"gradient"`
	LogoBase64      string           `json:"logoBase64"`
}

// Service normalizes style requests into validated render styles.
type Service struct {
	maxLogoBytes int
}

// NewService creates a beautify service. maxLogoBytes bounds decoded logo input.
func NewService(maxLogoBytes int) *Service {
	if maxLogoBytes <= 0 {
		maxLogoBytes = 2 * 1024 * 1024
	}
	return &Service{maxLogoBytes: maxLogoBytes}
}

// Resolve validates and normalizes req into a qr.Style ready for rendering.
// It applies the named preset (if any) as defaults, then overlays explicit
// user values, clamps numeric ranges, and enforces contrast for scannability.
func (s *Service) Resolve(req StyleRequest) (qr.Style, error) {
	style := qr.Style{
		Foreground:      "#000000",
		Background:      "#FFFFFF",
		Shape:           qr.ShapeSquare,
		ErrorCorrection: "Q",
		Size:            defaultSize,
		BorderWidth:     defaultBorder,
	}

	if req.Preset != "" {
		preset, ok := findPreset(req.Preset)
		if !ok {
			return qr.Style{}, newValidationError("UNKNOWN_PRESET", "The requested preset does not exist")
		}
		style.Foreground = preset.Foreground
		style.Background = preset.Background
		style.Shape = preset.Shape
		style.Gradient = preset.Gradient
	}

	if req.Foreground != "" {
		fg, ok := normalizeHex(req.Foreground)
		if !ok {
			return qr.Style{}, newValidationError("INVALID_COLOR", "Foreground color must be a valid hex value")
		}
		style.Foreground = fg
	}

	if req.Background != "" {
		bg, ok := normalizeHex(req.Background)
		if !ok {
			return qr.Style{}, newValidationError("INVALID_COLOR", "Background color must be a valid hex value")
		}
		style.Background = bg
	}

	style.Transparent = req.Transparent

	if req.Shape != "" {
		switch qr.Shape(strings.ToLower(req.Shape)) {
		case qr.ShapeSquare:
			style.Shape = qr.ShapeSquare
		case qr.ShapeCircle:
			style.Shape = qr.ShapeCircle
		default:
			return qr.Style{}, newValidationError("INVALID_SHAPE", "Shape must be 'square' or 'circle'")
		}
	}

	if req.ErrorCorrection != "" {
		ec := strings.ToUpper(req.ErrorCorrection)
		switch ec {
		case "L", "M", "Q", "H":
			style.ErrorCorrection = ec
		default:
			return qr.Style{}, newValidationError("INVALID_ERROR_CORRECTION", "Error correction must be one of L, M, Q, H")
		}
	}

	if req.Size != 0 {
		style.Size = clampInt(req.Size, minSize, maxSize)
	}
	if req.BorderWidth != nil {
		style.BorderWidth = clampInt(*req.BorderWidth, 0, maxBorder)
	}

	if req.Gradient != nil {
		grad, err := s.resolveGradient(req.Gradient)
		if err != nil {
			return qr.Style{}, err
		}
		style.Gradient = grad
	}

	if req.LogoBase64 != "" {
		logo, err := s.decodeLogo(req.LogoBase64)
		if err != nil {
			return qr.Style{}, err
		}
		style.Logo = logo
		// A centered logo occludes modules; force the highest error
		// correction so the code remains scannable.
		style.ErrorCorrection = "H"
	}

	if err := s.checkScannability(style); err != nil {
		return qr.Style{}, err
	}

	return style, nil
}

func (s *Service) resolveGradient(g *GradientRequest) (*qr.Gradient, error) {
	if len(g.Stops) < 2 {
		return nil, newValidationError("INVALID_GRADIENT", "A gradient needs at least two color stops")
	}
	stops := make([]qr.GradientStop, 0, len(g.Stops))
	for _, st := range g.Stops {
		c, ok := normalizeHex(st.Color)
		if !ok {
			return nil, newValidationError("INVALID_COLOR", "Gradient stop color must be a valid hex value")
		}
		stops = append(stops, qr.GradientStop{Offset: st.Offset, Color: c})
	}
	return &qr.Gradient{Angle: g.Angle, Stops: stops}, nil
}

// decodeLogo accepts a raw base64 string or a data URL and returns the decoded
// image after validating format (PNG/JPEG) and byte size.
func (s *Service) decodeLogo(input string) (image.Image, error) {
	payload := input
	if idx := strings.Index(payload, "base64,"); idx != -1 {
		payload = payload[idx+len("base64,"):]
	}
	payload = strings.TrimSpace(payload)

	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, newValidationError("INVALID_LOGO", "Logo must be valid base64-encoded image data")
	}
	if len(raw) > s.maxLogoBytes {
		return nil, newValidationError("LOGO_TOO_LARGE", "Logo image exceeds the maximum allowed size")
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, newValidationError("INVALID_LOGO", "Logo must be a valid PNG or JPEG image")
	}
	if format != "png" && format != "jpeg" {
		return nil, newValidationError("UNSUPPORTED_LOGO_FORMAT", "Logo must be a PNG or JPEG image")
	}
	return img, nil
}

// checkScannability rejects color combinations too low in contrast to scan.
// Transparent backgrounds are skipped since the surface color is unknown.
func (s *Service) checkScannability(style qr.Style) error {
	if style.Transparent {
		return nil
	}

	colors := []string{style.Foreground}
	if style.Gradient != nil {
		colors = colors[:0]
		for _, st := range style.Gradient.Stops {
			colors = append(colors, st.Color)
		}
	}

	for _, fg := range colors {
		if contrastRatio(fg, style.Background) < minContrastRatio {
			return newValidationError(
				"LOW_CONTRAST",
				"Foreground and background colors are too similar; increase contrast so the code can be scanned",
			)
		}
	}
	return nil
}

func normalizeHex(hex string) (string, bool) {
	h := strings.TrimSpace(hex)
	if !hexColorPattern.MatchString(h) {
		return "", false
	}
	h = strings.TrimPrefix(h, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	return "#" + strings.ToUpper(h), true
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// contrastRatio returns the WCAG contrast ratio (1.0 .. 21.0) between two hex
// colors. Returns 1.0 (no contrast) if either color cannot be parsed.
func contrastRatio(hexA, hexB string) float64 {
	la, okA := relativeLuminance(hexA)
	lb, okB := relativeLuminance(hexB)
	if !okA || !okB {
		return 1.0
	}
	lighter, darker := la, lb
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

func relativeLuminance(hex string) (float64, bool) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) != 6 {
		return 0, false
	}
	channels := [3]float64{}
	for i := 0; i < 3; i++ {
		var v int
		for _, c := range h[i*2 : i*2+2] {
			d := hexDigit(c)
			if d < 0 {
				return 0, false
			}
			v = v*16 + d
		}
		channels[i] = linearize(float64(v) / 255.0)
	}
	return 0.2126*channels[0] + 0.7152*channels[1] + 0.0722*channels[2], true
}

func linearize(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func hexDigit(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'a' && r <= 'f':
		return int(r-'a') + 10
	case r >= 'A' && r <= 'F':
		return int(r-'A') + 10
	default:
		return -1
	}
}
