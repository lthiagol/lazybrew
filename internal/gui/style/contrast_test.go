package style

import (
	"image/color"
	"math"
	"testing"
)

// relLuminance is the WCAG relative luminance of a colour.
func relLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	lin := func(v uint32) float64 {
		s := float64(v>>8) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrast is the WCAG contrast ratio of two colours, from 1 to 21.
func contrast(a, b color.Color) float64 {
	hi, lo := relLuminance(a), relLuminance(b)
	if lo > hi {
		hi, lo = lo, hi
	}
	return (hi + 0.05) / (lo + 0.05)
}

// contrastFloor is the least ratio a role may have on the theme's background:
// the body text 7:1, the status and secondary text 4.5:1, the muted text 3:1.
// Surface is not text and has no floor, but must differ from the background.
var contrastFloor = map[string]float64{
	"Text":    7,
	"Subtle":  4.5,
	"Accent":  4.5,
	"Success": 4.5,
	"Warning": 4.5,
	"Error":   4.5,
	"Info":    4.5,
	"Muted":   3,
}

// TestEveryThemeMeetsItsContrastFloor: each role's contrast with the theme's
// background meets its floor.
func TestEveryThemeMeetsItsContrastFloor(t *testing.T) {
	for _, name := range registryNames() {
		th, _ := ThemeByName(name)
		roles := roleSet(th)
		for role, floor := range contrastFloor {
			ratio := contrast(roles[role], th.Background)
			if ratio < floor {
				t.Errorf("%s %s = %.2f:1 on the background, want at least %.1f:1", name, role, ratio, floor)
			}
		}
		if roles["Surface"] == th.Background {
			t.Errorf("%s: Surface is the background colour, want a distinct row colour", name)
		}
	}
}
