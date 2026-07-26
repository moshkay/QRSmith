package qr

import (
	"image/color"
	"strconv"
	"strings"
)

// parseHexRGBA parses a "#RRGGBB" or "#RGB" string into an opaque RGBA color.
func parseHexRGBA(hex string) (color.RGBA, bool) {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")

	switch len(h) {
	case 3: // shorthand #RGB -> #RRGGBB
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	case 6:
		// ok
	default:
		return color.RGBA{}, false
	}

	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return color.RGBA{}, false
	}

	return color.RGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 0xFF,
	}, true
}

func clamp01(f float64) float64 {
	switch {
	case f < 0:
		return 0
	case f > 1:
		return 1
	default:
		return f
	}
}
