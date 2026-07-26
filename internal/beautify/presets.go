package beautify

import "github.com/dojah/qrforge/internal/qr"

// Preset is a ready-made visual theme users can apply as a starting point.
type Preset struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Foreground  string       `json:"foreground"`
	Background  string       `json:"background"`
	Shape       qr.Shape     `json:"shape"`
	Gradient    *qr.Gradient `json:"gradient,omitempty"`
}

// presets is the built-in theme catalogue, keyed by ID for quick lookup.
var presets = []Preset{
	{
		ID:          "classic",
		Name:        "Classic",
		Description: "Timeless black on white. Maximum scannability.",
		Foreground:  "#000000",
		Background:  "#FFFFFF",
		Shape:       qr.ShapeSquare,
	},
	{
		ID:          "midnight",
		Name:        "Midnight",
		Description: "Deep navy modules on a soft off-white field.",
		Foreground:  "#0B1F3A",
		Background:  "#F5F7FA",
		Shape:       qr.ShapeSquare,
	},
	{
		ID:          "ocean",
		Name:        "Ocean",
		Description: "Teal-to-blue gradient with rounded dots.",
		Foreground:  "#0EA5E9",
		Background:  "#FFFFFF",
		Shape:       qr.ShapeCircle,
		Gradient: &qr.Gradient{
			Angle: 45,
			Stops: []qr.GradientStop{
				{Offset: 0, Color: "#06B6D4"},
				{Offset: 1, Color: "#2563EB"},
			},
		},
	},
	{
		ID:          "sunset",
		Name:        "Sunset",
		Description: "Warm orange-to-pink gradient dots.",
		Foreground:  "#F97316",
		Background:  "#FFFFFF",
		Shape:       qr.ShapeCircle,
		Gradient: &qr.Gradient{
			Angle: 90,
			Stops: []qr.GradientStop{
				{Offset: 0, Color: "#F97316"},
				{Offset: 1, Color: "#DB2777"},
			},
		},
	},
	{
		ID:          "forest",
		Name:        "Forest",
		Description: "Rich green on a pale mint background.",
		Foreground:  "#166534",
		Background:  "#F0FDF4",
		Shape:       qr.ShapeSquare,
	},
	{
		ID:          "neon",
		Name:        "Neon",
		Description: "Electric green on near-black. High energy.",
		Foreground:  "#39FF14",
		Background:  "#0A0A0A",
		Shape:       qr.ShapeCircle,
	},
	{
		ID:          "grape",
		Name:        "Grape",
		Description: "Violet-to-indigo gradient with rounded dots.",
		Foreground:  "#7C3AED",
		Background:  "#FFFFFF",
		Shape:       qr.ShapeCircle,
		Gradient: &qr.Gradient{
			Angle: 135,
			Stops: []qr.GradientStop{
				{Offset: 0, Color: "#8B5CF6"},
				{Offset: 1, Color: "#4338CA"},
			},
		},
	},
}

// Presets returns a copy of the built-in preset catalogue.
func Presets() []Preset {
	out := make([]Preset, len(presets))
	copy(out, presets)
	return out
}

func findPreset(id string) (Preset, bool) {
	for _, p := range presets {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}
