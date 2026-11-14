package table

// LineWeight enum value (code 370)
var LineWidth = map[int]float64{
	0:   0.00,
	5:   0.05,
	9:   0.09,
	13:  0.13,
	15:  0.15,
	18:  0.18,
	20:  0.20,
	25:  0.25,
	30:  0.30,
	35:  0.35,
	40:  0.40,
	50:  0.50,
	53:  0.53,
	60:  0.60,
	70:  0.70,
	80:  0.80,
	90:  0.90,
	100: 1.00,
	106: 1.06,
	120: 1.20,
	140: 1.40,
	158: 1.58,
	200: 2.00,
	211: 2.11,
}

// Special lineweight values (code 370).
const (
	LineWeightByLayer = -1
	LineWeightByBlock = -2
	LineWeightDefault = -3
)

// NearestLineWeight returns a valid lineweight (code 370, 1/100 mm) for w:
// w itself if it is valid or one of the special values (BYLAYER, BYBLOCK,
// DEFAULT), otherwise the smallest valid value greater than w (211 at most).
// Other negative values are treated as BYLAYER.
func NearestLineWeight(w int) int {
	switch {
	case w == LineWeightByLayer || w == LineWeightByBlock || w == LineWeightDefault:
		return w
	case w < 0:
		return LineWeightByLayer
	case w > 211:
		return 211
	}
	if _, ok := LineWidth[w]; ok {
		return w
	}
	best := 211
	for k := range LineWidth {
		if k > w && k < best {
			best = k
		}
	}
	return best
}
