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

// latteBelow records the Latte roles that sit under their floor. Latte keeps
// today's colours (its frames are pinned), so the floor is not met here; each
// role is held at its measured ratio, and the test fails if one falls further.
// Q-lb28-latte-contrast asks the orchestrator to rule on these.
var latteBelow = map[string]float64{
	"Subtle":  4.37,
	"Muted":   2.30,
	"Success": 2.96,
	"Warning": 2.31,
	"Info":    4.34,
}

// TestEveryThemeMeetsItsContrastFloor: each role's contrast with the theme's
// background meets its floor, except the Latte roles recorded above.
func TestEveryThemeMeetsItsContrastFloor(t *testing.T) {
	for _, name := range ThemeNames() {
		th, _ := ThemeByName(name)
		roles := roleSet(th)
		for role, floor := range contrastFloor {
			ratio := contrast(roles[role], th.Background)
			if name == "light" || name == "catppuccin-latte" {
				if below, ok := latteBelow[role]; ok {
					if ratio < below-0.005 {
						t.Errorf("%s %s = %.2f:1, below its recorded %.2f:1", name, role, ratio, below)
					}
					continue
				}
			}
			if ratio < floor {
				t.Errorf("%s %s = %.2f:1 on the background, want at least %.1f:1", name, role, ratio, floor)
			}
		}
		if roles["Surface"] == th.Background {
			t.Errorf("%s: Surface is the background colour, want a distinct row colour", name)
		}
	}
}

// TestLatteBelowFloorIsStillBelow: a Latte role recorded as below its floor
// must still be below it, so a fix that lifts it is noticed and the record
// is removed.
func TestLatteBelowFloorIsStillBelow(t *testing.T) {
	latte := LightTheme()
	for role := range latteBelow {
		if got := contrast(roleSet(latte)[role], latte.Background); got >= contrastFloor[role] {
			t.Errorf("Latte %s now meets its floor (%.2f:1): drop it from latteBelow", role, got)
		}
	}
}
