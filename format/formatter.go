// DXF Output Formatter
package format

import (
	"io"
)

// Formatter controls output format.
type Formatter interface {
	Reset()
	WriteTo(w io.Writer) (int64, error)
	SetPrecision(p int)
	Output() string
	String(num int, val string) string
	Hex(num int, h int) string
	Int(num int, val int) string
	Float(num int, val float64) string
	WriteString(num int, val string)
	WriteHex(num int, h int)
	WriteInt(num int, val int)
	WriteFloat(num int, val float64)
}
