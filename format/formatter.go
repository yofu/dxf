package format

import (
	"bytes"
	"fmt"
	"io"
)

type Formatter struct {
	buffer bytes.Buffer
	float  string
}

func New() *Formatter {
	var b bytes.Buffer
	return &Formatter{
		buffer: b,
		float:  "%.6f",
	}
}

func (f *Formatter) Reset() {
	f.buffer.Reset()
}

func (f *Formatter) WriteTo(w io.Writer) (int64, error) {
	return f.buffer.WriteTo(w)
}

func (f *Formatter) SetPrecision(p int) {
	f.float = fmt.Sprintf("%%.%df", p)
}

func (f *Formatter) Output() string {
	rtn := f.buffer.String()
	f.buffer.Reset()
	return rtn
}

func (f *Formatter) String(num int, val string) string {
	return fmt.Sprintf("%d\n%s\n", num, val)
}

func (f *Formatter) Hex(num int, h int) string {
	return fmt.Sprintf("%d\n%X\n", num, h)
}

func (f *Formatter) Int(num int, val int) string {
	return fmt.Sprintf("%d\n%d\n", num, val)
}

func (f *Formatter) Float(num int, val float64) string {
	return fmt.Sprintf(fmt.Sprintf("%d\n%s\n", num, f.float), val)
}

func (f *Formatter) WriteString(num int, val string) {
	f.buffer.WriteString(f.String(num, val))
}

func (f *Formatter) WriteHex(num int, h int) {
	f.buffer.WriteString(f.Hex(num, h))
}

func (f *Formatter) WriteInt(num int, val int) {
	f.buffer.WriteString(f.Int(num, val))
}

func (f *Formatter) WriteFloat(num int, val float64) {
	f.buffer.WriteString(f.Float(num, val))
}
