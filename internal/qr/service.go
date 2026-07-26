// Package qr is the core rendering engine. It turns text content plus a
// validated Style into a PNG image, supporting colors, gradients, shapes,
// borders, a centered logo, and a fixed output size.
package qr

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"

	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	xdraw "golang.org/x/image/draw"
)

// Shape describes how each QR module is drawn.
type Shape string

const (
	ShapeSquare Shape = "square"
	ShapeCircle Shape = "circle"
)

// GradientStop is one color stop along a linear gradient.
type GradientStop struct {
	Offset float64 `json:"offset"` // 0.0 .. 1.0
	Color  string  `json:"color"`  // #RRGGBB
}

// Gradient describes a linear foreground gradient.
type Gradient struct {
	Angle float64        `json:"angle"` // degrees: 0=right, 90=up, 180=left, 270=down
	Stops []GradientStop `json:"stops"`
}

// Style holds every visual option used to render a QR code. It is expected to
// be already validated/normalized by the beautify service.
type Style struct {
	Foreground      string      // #RRGGBB
	Background      string      // #RRGGBB (ignored if Transparent)
	Transparent     bool        // transparent background
	BorderWidth     int         // quiet-zone width in pixels
	Shape           Shape       // square or circle modules
	ErrorCorrection string      // "L", "M", "Q", "H"
	Size            int         // output image width/height in pixels
	Gradient        *Gradient   // optional foreground gradient
	Logo            image.Image // optional centered logo
}

// Generator renders QR codes. It is stateless and safe for concurrent use.
type Generator struct{}

// NewGenerator returns a new QR generator.
func NewGenerator() *Generator { return &Generator{} }

// nopWriteCloser adapts an io.Writer (a bytes.Buffer) to io.WriteCloser so the
// standard writer can write into an in-memory buffer instead of a file.
type nopWriteCloser struct{ *bytes.Buffer }

func (nopWriteCloser) Close() error { return nil }

// Render encodes content and returns the resulting PNG image bytes.
func (g *Generator) Render(content string, style Style) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("qr: empty content")
	}

	code, err := qrcode.NewWith(content, errorCorrectionOption(style.ErrorCorrection))
	if err != nil {
		return nil, fmt.Errorf("qr: encode content: %w", err)
	}

	opts := buildImageOptions(style)

	buf := &bytes.Buffer{}
	w := standard.NewWithWriter(nopWriteCloser{buf}, opts...)
	if err := code.Save(w); err != nil {
		return nil, fmt.Errorf("qr: render image: %w", err)
	}

	if style.Size <= 0 {
		return buf.Bytes(), nil
	}
	return resizePNG(buf.Bytes(), style.Size)
}

func buildImageOptions(style Style) []standard.ImageOption {
	// A generous per-module width keeps the base image sharp before it is
	// resized down to the requested output size.
	const baseModuleWidth = 16

	opts := []standard.ImageOption{
		standard.WithQRWidth(baseModuleWidth),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		standard.WithBorderWidth(style.BorderWidth),
		standard.WithFgColorRGBHex(style.Foreground),
	}

	if style.Transparent {
		opts = append(opts, standard.WithBgTransparent())
	} else {
		opts = append(opts, standard.WithBgColorRGBHex(style.Background))
	}

	if style.Shape == ShapeCircle {
		opts = append(opts, standard.WithCircleShape())
	}

	if g := style.Gradient; g != nil && len(g.Stops) > 0 {
		if grad := buildGradient(g); grad != nil {
			opts = append(opts, standard.WithFgGradient(grad))
		}
	}

	if style.Logo != nil {
		opts = append(opts,
			standard.WithLogoImage(style.Logo),
			standard.WithLogoSizeMultiplier(5), // logo <= 1/5 of QR width
		)
	}

	return opts
}

func buildGradient(g *Gradient) *standard.LinearGradient {
	stops := make([]standard.ColorStop, 0, len(g.Stops))
	for _, s := range g.Stops {
		c, ok := parseHexRGBA(s.Color)
		if !ok {
			continue
		}
		stops = append(stops, standard.ColorStop{T: clamp01(s.Offset), Color: c})
	}
	if len(stops) == 0 {
		return nil
	}
	return standard.NewGradient(g.Angle, stops...)
}

// resizePNG scales a PNG to a size x size square using nearest-neighbor
// interpolation, which keeps QR module edges crisp and scannable.
func resizePNG(src []byte, size int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("qr: decode for resize: %w", err)
	}

	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	xdraw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)

	out := &bytes.Buffer{}
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(out, dst); err != nil {
		return nil, fmt.Errorf("qr: encode resized: %w", err)
	}
	return out.Bytes(), nil
}

// errorCorrectionOption converts an "L"/"M"/"Q"/"H" string into the library's
// error correction encode option, defaulting to Quart (25%) for unknown input.
// Higher levels recover from more damage/occlusion (e.g. a centered logo).
func errorCorrectionOption(level string) qrcode.EncodeOption {
	switch strings.ToUpper(level) {
	case "L":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case "M":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	case "H":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	}
}
