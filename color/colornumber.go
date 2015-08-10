package color

type ColorNumber uint8

// 1 - 9
const (
	Red ColorNumber = iota + 1
	Yellow
	Green
	Cyan
	Blue
	Magenta
	White
	Grey128
	Grey192
)

// 250 - 255
const (
	Grey51 = iota + 250
	Grey91
	Grey132
	Grey173
	Grey214
	Grey255
)
