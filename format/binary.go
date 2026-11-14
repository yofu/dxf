package format

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// BinarySentinel is the sentinel at the beginning of binary DXF files.
const BinarySentinel = "AutoCAD Binary DXF\r\n\x1a\x00"

// ValueType is the type of the value of a group code in binary DXF.
type ValueType int

// Value types of group codes.
const (
	TypeString ValueType = iota
	TypeDouble
	TypeInt16
	TypeInt32
	TypeInt64
	TypeBool
	TypeBinary
)

// GroupCodeType returns the value type of group code in binary DXF
// (DXF reference, "Group Code Value Types").
func GroupCodeType(code int) ValueType {
	switch {
	case code >= 10 && code <= 59,
		code >= 110 && code <= 149,
		code >= 210 && code <= 239,
		code >= 460 && code <= 469,
		code >= 1010 && code <= 1059:
		return TypeDouble
	case code >= 60 && code <= 79,
		code >= 170 && code <= 179,
		code >= 270 && code <= 289,
		code >= 370 && code <= 389,
		code >= 400 && code <= 409,
		code >= 1060 && code <= 1070:
		return TypeInt16
	case code >= 90 && code <= 99,
		code >= 420 && code <= 429,
		code >= 440 && code <= 459,
		code == 1071:
		return TypeInt32
	case code >= 160 && code <= 169:
		return TypeInt64
	case code >= 290 && code <= 299:
		return TypeBool
	case code >= 310 && code <= 319, code == 1004:
		return TypeBinary
	default:
		return TypeString
	}
}

// Binary is Formatter for binary DXF (AC1012 or later: 2-byte group codes).
//
// Values are written in the type of their group code, whatever Write*
// method is used: e.g. WriteString(10, "1.5") writes a double.
type Binary struct {
	buffer  bytes.Buffer
	encode  func(string) string
	err     error
	version string
}

// NewBinary creates a new Binary formatter.
func NewBinary() *Binary {
	f := &Binary{}
	f.Reset()
	return f
}

// SetEncoder sets the function converting string values to the code page
// of the file (nil: written as they are).
func (f *Binary) SetEncoder(enc func(string) string) {
	f.encode = enc
}

// Reset resets the buffer; it starts with the sentinel.
func (f *Binary) Reset() {
	f.buffer.Reset()
	f.buffer.WriteString(BinarySentinel)
	f.err = nil
}

// Err returns the first error of conversion of a value to the type of its
// group code (e.g. WriteString(10, "abc")), or nil.
func (f *Binary) Err() error {
	return f.err
}

// WriteTo writes data stored in the buffer to w.
func (f *Binary) WriteTo(w io.Writer) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.buffer.WriteTo(w)
}

// SetPrecision does nothing: doubles are written as they are.
func (f *Binary) SetPrecision(p int) {}

// Output outputs data stored in the buffer and resets it.
func (f *Binary) Output() string {
	rtn := f.buffer.String()
	f.Reset()
	return rtn
}

func (f *Binary) setErr(num int, val string, err error) {
	if f.err == nil {
		f.err = fmt.Errorf("binary dxf: group code %d: value %q: %w", num, val, err)
	}
}

// pair encodes group code num and value given as text.
func (f *Binary) pair(num int, val string) string {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, int16(num))
	v := strings.TrimSpace(val)
	switch GroupCodeType(num) {
	case TypeDouble:
		x, err := strconv.ParseFloat(v, 64)
		if err != nil {
			f.setErr(num, val, err)
		}
		binary.Write(&b, binary.LittleEndian, x)
	case TypeInt16:
		x, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			f.setErr(num, val, err)
		}
		binary.Write(&b, binary.LittleEndian, int16(x))
	case TypeInt32:
		x, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			f.setErr(num, val, err)
		}
		binary.Write(&b, binary.LittleEndian, int32(x))
	case TypeInt64:
		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			f.setErr(num, val, err)
		}
		binary.Write(&b, binary.LittleEndian, x)
	case TypeBool:
		x, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			f.setErr(num, val, err)
		}
		if x != 0 {
			b.WriteByte(1)
		} else {
			b.WriteByte(0)
		}
	case TypeBinary:
		data, err := hex.DecodeString(v)
		if err != nil {
			f.setErr(num, val, err)
		}
		// at most 255 bytes per group; longer data is split
		for first := true; first || len(data) > 0; first = false {
			n := len(data)
			if n > 255 {
				n = 255
			}
			if !first {
				binary.Write(&b, binary.LittleEndian, int16(num))
			}
			b.WriteByte(byte(n))
			b.Write(data[:n])
			data = data[n:]
		}
	default:
		if f.encode != nil {
			val = f.encode(val)
		}
		b.WriteString(val)
		b.WriteByte(0)
	}
	return b.String()
}

// String outputs given code & string in binary DXF.
func (f *Binary) String(num int, val string) string {
	return f.pair(num, val)
}

// Hex outputs given code & hex (handle) in binary DXF.
func (f *Binary) Hex(num int, h int) string {
	return f.pair(num, fmt.Sprintf("%X", h))
}

// Int outputs given code & int in binary DXF.
func (f *Binary) Int(num int, val int) string {
	return f.pair(num, strconv.Itoa(val))
}

// Float outputs given code & floating point in binary DXF.
func (f *Binary) Float(num int, val float64) string {
	if GroupCodeType(num) == TypeDouble {
		var b bytes.Buffer
		binary.Write(&b, binary.LittleEndian, int16(num))
		binary.Write(&b, binary.LittleEndian, val)
		return b.String()
	}
	if math.Trunc(val) == val {
		return f.pair(num, strconv.FormatInt(int64(val), 10))
	}
	return f.pair(num, strconv.FormatFloat(val, 'g', -1, 64))
}

// WriteString appends string data to the buffer.
func (f *Binary) WriteString(num int, val string) {
	f.buffer.WriteString(f.String(num, val))
}

// WriteHex appends hex data to the buffer.
func (f *Binary) WriteHex(num int, h int) {
	f.buffer.WriteString(f.Hex(num, h))
}

// WriteInt appends int data to the buffer.
func (f *Binary) WriteInt(num int, val int) {
	f.buffer.WriteString(f.Int(num, val))
}

// WriteFloat appends floating point data to the buffer.
func (f *Binary) WriteFloat(num int, val float64) {
	f.buffer.WriteString(f.Float(num, val))
}

var _ Formatter = &Binary{}
