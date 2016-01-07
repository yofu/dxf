package format

import (
	"bytes"
	"fmt"
	"io"
)

// ASCII is Formatter for ASCII format.
type ASCII struct {
	buffer bytes.Buffer
	float  string
}

// NewASCII creates a new ASCII formatter.
func NewASCII() *ASCII {
	var b bytes.Buffer
	return &ASCII{
		buffer: b,
		float:  "%.6f",
	}
}

// Reset resets the buffer.
func (f *ASCII) Reset() {
	f.buffer.Reset()
}

// WriteTo writes data stored in the buffer to w.
func (f *ASCII) WriteTo(w io.Writer) (int64, error) {
	return f.buffer.WriteTo(w)
}

// SetPrecision sets precision part for outputting floating point values.
func (f *ASCII) SetPrecision(p int) {
	f.float = fmt.Sprintf("%%.%df", p)
}

// Output outputs data stored in the buffer in DXF format.
func (f *ASCII) Output() string {
	rtn := f.buffer.String()
	f.buffer.Reset()
	return rtn
}

// String outputs given code & string in DXF format.
func (f *ASCII) String(num int, val string) string {
	return fmt.Sprintf("%d\n%s\n", num, val)
}

// Hex outputs given code & hex in DXF format.
// It is used for outputting handles.
func (f *ASCII) Hex(num int, h int) string {
	return fmt.Sprintf("%d\n%X\n", num, h)
}

// Int outputs given code & int in DXF format.
func (f *ASCII) Int(num int, val int) string {
	return fmt.Sprintf("%d\n%d\n", num, val)
}

// Float outputs given code & floating point in DXF format.
func (f *ASCII) Float(num int, val float64) string {
	return fmt.Sprintf(fmt.Sprintf("%d\n%s\n", num, f.float), val)
}

// WriteString appends string data to the buffer.
func (f *ASCII) WriteString(num int, val string) {
	f.buffer.WriteString(f.String(num, val))
}

// WriteHex appends hex data to the buffer.
func (f *ASCII) WriteHex(num int, h int) {
	f.buffer.WriteString(f.Hex(num, h))
}

// WriteInt appends int data to the buffer.
func (f *ASCII) WriteInt(num int, val int) {
	f.buffer.WriteString(f.Int(num, val))
}

// WriteFloat appends floating point data to the buffer.
func (f *ASCII) WriteFloat(num int, val float64) {
	f.buffer.WriteString(f.Float(num, val))
}
