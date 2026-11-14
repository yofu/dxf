package entity

import (
	"math"
	"strings"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/format"
)

// Gradient names of HATCH (code 470).
const (
	GRADIENT_LINEAR           = "LINEAR"
	GRADIENT_CYLINDER         = "CYLINDER"
	GRADIENT_INVCYLINDER      = "INVCYLINDER"
	GRADIENT_SPHERICAL        = "SPHERICAL"
	GRADIENT_INVSPHERICAL     = "INVSPHERICAL"
	GRADIENT_HEMISPHERICAL    = "HEMISPHERICAL"
	GRADIENT_INVHEMISPHERICAL = "INVHEMISPHERICAL"
	GRADIENT_CURVED           = "CURVED"
	GRADIENT_INVCURVED        = "INVCURVED"
)

// GradientColor is a color of a HATCH gradient.
type GradientColor struct {
	Value float64 // 463: position of the color (0 or 1)
	ACI   int     // 63: color number (0: not written)
	RGB   int     // 421: true color value (0xRRGGBB)
}

// HatchGradient is the gradient fill of HATCH (AC1018 or later).
// A gradient hatch is a solid hatch ("SOLID", 70 = 1); older versions and
// applications that do not support gradients show it as a solid fill.
type HatchGradient struct {
	Name     string          // 470: GRADIENT_LINEAR, ...
	Rotation float64         // 460: rotation angle in degrees (written in radians)
	Centered float64         // 461: 0 = centered, 1 = shifted (AutoCAD writes 0 or 1)
	OneColor bool            // 452: one-color gradient
	Tint     float64         // 462: tint of a one-color gradient (0: dark, 1: light)
	Colors   []GradientColor // 453, 463, 63, 421
}

func (g *HatchGradient) format(f format.Formatter) {
	f.WriteInt(450, 1) // gradient fill
	f.WriteInt(451, 0) // reserved
	f.WriteFloat(460, g.Rotation*math.Pi/180.0)
	f.WriteFloat(461, g.Centered)
	f.WriteInt(452, boolInt(g.OneColor))
	f.WriteFloat(462, g.Tint)
	f.WriteInt(453, len(g.Colors))
	for _, c := range g.Colors {
		f.WriteFloat(463, c.Value)
		if c.ACI != 0 {
			f.WriteInt(63, c.ACI)
		}
		f.WriteInt(421, c.RGB)
	}
	f.WriteString(470, g.Name)
}

// SetGradient sets a two-color gradient fill: name is one of GRADIENT_*,
// rgb1 and rgb2 are true color values (color.RGB) and rotation is in degrees.
// The hatch becomes a solid hatch, which is how it is shown by AutoCAD
// before AC1018. The gradient is written for AC1018 or later.
func (h *Hatch) SetGradient(name string, rgb1, rgb2 int, rotation float64) *HatchGradient {
	h.SetSolid()
	h.Gradient = &HatchGradient{
		Name:     strings.ToUpper(name),
		Rotation: rotation,
		Colors: []GradientColor{
			{Value: 0, RGB: rgb1},
			{Value: 1, RGB: rgb2},
		},
	}
	return h.Gradient
}

// SetOneColorGradient sets a one-color gradient fill with tint
// (0: dark to 1: light). The second color is the tinted color as written by
// AutoCAD, approximated as black (tint 0) or white (tint 1) blended by tint.
func (h *Hatch) SetOneColorGradient(name string, rgb int, tint, rotation float64) *HatchGradient {
	t := math.Max(0, math.Min(1, tint))
	v := uint8(math.Round(255 * t))
	g := h.SetGradient(name, rgb, color.RGB(v, v, v), rotation)
	g.OneColor = true
	g.Tint = t
	return g
}
