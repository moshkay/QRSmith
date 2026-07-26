package qr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"math"
	"strings"

	qrcode "github.com/yeqown/go-qrcode/v2"
)

// baseModulePx mirrors the per-module pixel width used by the raster renderer,
// so a pixel-based BorderWidth maps to the same quiet zone (in modules) here.
const baseModulePx = 16

// RenderSVG encodes content and returns a scalable SVG document. SVG output is
// resolution-independent, ideal for print. It honors colors, gradient, module
// shape, quiet zone, and an optional centered logo.
func (g *Generator) RenderSVG(content string, style Style) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("qr: empty content")
	}

	code, err := qrcode.NewWith(content, errorCorrectionOption(style.ErrorCorrection))
	if err != nil {
		return nil, fmt.Errorf("qr: encode content: %w", err)
	}

	w := &svgWriter{style: style}
	if err := code.Save(w); err != nil {
		return nil, fmt.Errorf("qr: render svg: %w", err)
	}
	return w.buf.Bytes(), nil
}

// svgWriter implements qrcode.Writer to emit an SVG document from the matrix.
type svgWriter struct {
	style Style
	buf   bytes.Buffer
}

func (w *svgWriter) Close() error { return nil }

func (w *svgWriter) Write(mat qrcode.Matrix) error {
	bitmap := mat.Bitmap()
	n := mat.Width()

	quiet := w.style.BorderWidth / baseModulePx
	if quiet < 0 {
		quiet = 0
	}
	total := n + quiet*2

	size := w.style.Size
	if size <= 0 {
		size = total * baseModulePx
	}

	var b strings.Builder
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`,
		size, size, total, total)

	fill := w.fillRef(&b) // may inject a <defs> gradient and returns the fill value

	if !w.style.Transparent {
		fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`, total, total, escapeAttr(w.style.Background))
	}

	if w.style.Shape == ShapeCircle {
		w.writeCircles(&b, bitmap, n, quiet, fill)
	} else {
		w.writeSquares(&b, bitmap, n, quiet, fill)
	}

	if w.style.Logo != nil {
		w.writeLogo(&b, total)
	}

	b.WriteString(`</svg>`)
	w.buf.WriteString(b.String())
	return nil
}

// fillRef returns the fill value for modules. When a gradient is configured it
// writes a <defs><linearGradient> block and returns a url(#id) reference.
func (w *svgWriter) fillRef(b *strings.Builder) string {
	g := w.style.Gradient
	if g == nil || len(g.Stops) == 0 {
		return escapeAttr(w.style.Foreground)
	}

	// Convert angle (0=right, 90=up) to objectBoundingBox coordinates. SVG's
	// y-axis points down, so "up" is negative dy.
	rad := g.Angle * math.Pi / 180.0
	dx := math.Cos(rad)
	dy := -math.Sin(rad)
	x1 := 0.5 - dx/2
	y1 := 0.5 - dy/2
	x2 := 0.5 + dx/2
	y2 := 0.5 + dy/2

	b.WriteString(`<defs><linearGradient id="fg" gradientUnits="objectBoundingBox"`)
	fmt.Fprintf(b, ` x1="%.4f" y1="%.4f" x2="%.4f" y2="%.4f">`, x1, y1, x2, y2)
	for _, s := range g.Stops {
		fmt.Fprintf(b, `<stop offset="%.4f" stop-color="%s"/>`, clamp01(s.Offset), escapeAttr(s.Color))
	}
	b.WriteString(`</linearGradient></defs>`)
	return "url(#fg)"
}

func (w *svgWriter) writeSquares(b *strings.Builder, bitmap [][]bool, n, quiet int, fill string) {
	// Combine all dark modules into a single path for a compact document.
	var path strings.Builder
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if bitmap[y][x] {
				fmt.Fprintf(&path, "M%d %dh1v1h-1z", x+quiet, y+quiet)
			}
		}
	}
	fmt.Fprintf(b, `<path d="%s" fill="%s"/>`, path.String(), fill)
}

func (w *svgWriter) writeCircles(b *strings.Builder, bitmap [][]bool, n, quiet int, fill string) {
	fmt.Fprintf(b, `<g fill="%s">`, fill)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if bitmap[y][x] {
				cx := float64(x+quiet) + 0.5
				cy := float64(y+quiet) + 0.5
				fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="0.48"/>`, cx, cy)
			}
		}
	}
	b.WriteString(`</g>`)
}

// writeLogo embeds the logo (re-encoded as a PNG data URI) centered on the code,
// with a white rounded backing plate so it stays visually distinct.
func (w *svgWriter) writeLogo(b *strings.Builder, total int) {
	var logoBuf bytes.Buffer
	if err := png.Encode(&logoBuf, w.style.Logo); err != nil {
		return
	}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(logoBuf.Bytes())

	logoSize := float64(total) * 0.2
	pos := (float64(total) - logoSize) / 2
	pad := logoSize * 0.12

	fmt.Fprintf(b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="%.2f" fill="#FFFFFF"/>`,
		pos-pad, pos-pad, logoSize+2*pad, logoSize+2*pad, logoSize*0.15)
	fmt.Fprintf(b, `<image x="%.2f" y="%.2f" width="%.2f" height="%.2f" href="%s" preserveAspectRatio="xMidYMid meet"/>`,
		pos, pos, logoSize, logoSize, dataURI)
}

// escapeAttr escapes the minimal set of characters unsafe in an SVG attribute.
func escapeAttr(s string) string {
	r := strings.NewReplacer(`&`, "&amp;", `"`, "&quot;", `<`, "&lt;", `>`, "&gt;")
	return r.Replace(s)
}
