package color

import "math"

// RGB returns a true color value (code 420: 0x00RRGGBB).
func RGB(r, g, b uint8) int {
	return int(r)<<16 | int(g)<<8 | int(b)
}

// SplitRGB returns the red, green and blue components of a true color value.
func SplitRGB(c int) (r, g, b uint8) {
	return uint8(c >> 16), uint8(c >> 8), uint8(c)
}

// Nearest returns the color number (1-255) nearest to the RGB color.
func Nearest(r, g, b uint8) ColorNumber {
	best, bestDist := 7, -1
	for i := 1; i < len(ColorRGB) && i < 256; i++ {
		c := ColorRGB[i]
		dr, dg, db := int(r)-int(c[0]), int(g)-int(c[1]), int(b)-int(c[2])
		dist := dr*dr + dg*dg + db*db
		if bestDist < 0 || dist < bestDist {
			best, bestDist = i, dist
			if dist == 0 {
				break
			}
		}
	}
	return ColorNumber(best)
}

// Transparency values (code 440).
const (
	TransparencyByBlock = 0x01000000
	transparencyAlpha   = 0x02000000
)

// EncodeTransparency returns the value of code 440 for a transparency
// (0: opaque, 1: fully transparent).
func EncodeTransparency(t float64) int {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return transparencyAlpha | int(math.Round((1-t)*255))
}

// DecodeTransparency returns the transparency (0: opaque, 1: fully
// transparent) of a value of code 440; ok is false for BYLAYER and BYBLOCK.
func DecodeTransparency(v int) (float64, bool) {
	if v&transparencyAlpha == 0 {
		return 0, false
	}
	return 1 - float64(v&0xFF)/255, true
}
