package format

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestGroupCodeType(t *testing.T) {
	cases := map[int]ValueType{
		0: TypeString, 5: TypeString, 8: TypeString, 10: TypeDouble, 40: TypeDouble, 62: TypeInt16,
		70: TypeInt16, 90: TypeInt32, 100: TypeString, 105: TypeString, 140: TypeDouble, 160: TypeInt64,
		210: TypeDouble, 271: TypeInt16, 281: TypeInt16, 290: TypeBool, 310: TypeBinary, 330: TypeString,
		370: TypeInt16, 390: TypeString, 420: TypeInt32, 999: TypeString, 1000: TypeString, 1004: TypeBinary,
		1005: TypeString, 1040: TypeDouble, 1070: TypeInt16, 1071: TypeInt32,
	}
	for code, want := range cases {
		if got := GroupCodeType(code); got != want {
			t.Errorf("GroupCodeType(%d) = %v, want %v", code, got, want)
		}
	}
}

func le(v interface{}) []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestBinaryFormatter(t *testing.T) {
	f := NewBinary()
	f.WriteString(0, "SECTION")
	f.WriteFloat(10, 1.5)
	f.WriteString(20, "-2.25") // converted to the type of the group code
	f.WriteInt(70, -3)
	f.WriteInt(90, 100000)
	f.WriteHex(5, 0x2F)
	f.WriteInt(290, 1)
	f.WriteString(310, "0102FF")
	f.WriteFloat(62, 7) // a float for an int16 code

	var want bytes.Buffer
	want.WriteString(BinarySentinel)
	want.Write(le(int16(0)))
	want.WriteString("SECTION\x00")
	want.Write(le(int16(10)))
	want.Write(le(1.5))
	want.Write(le(int16(20)))
	want.Write(le(-2.25))
	want.Write(le(int16(70)))
	want.Write(le(int16(-3)))
	want.Write(le(int16(90)))
	want.Write(le(int32(100000)))
	want.Write(le(int16(5)))
	want.WriteString("2F\x00")
	want.Write(le(int16(290)))
	want.WriteByte(1)
	want.Write(le(int16(310)))
	want.Write([]byte{3, 1, 2, 0xFF})
	want.Write(le(int16(62)))
	want.Write(le(int16(7)))

	if got := f.Output(); got != want.String() {
		t.Errorf("binary output:\n%q\nwant\n%q", got, want.String())
	}
	if f.Output() != BinarySentinel {
		t.Error("Output should reset the buffer to the sentinel")
	}
}

func TestBinaryLongChunkAndErrors(t *testing.T) {
	f := NewBinary()
	long := bytes.Repeat([]byte("AB"), 300) // 300 bytes
	f.WriteString(310, string(long))
	out := []byte(f.Output())[len(BinarySentinel):]
	// split into 255 + 45 bytes
	if len(out) != 2+1+255+2+1+45 || out[2] != 255 || out[2+1+255+2] != 45 {
		t.Errorf("long binary chunk: %d bytes", len(out))
	}

	f.WriteString(10, "abc")
	if f.Err() == nil {
		t.Error("invalid double should be an error")
	}
	if _, err := f.WriteTo(&bytes.Buffer{}); err == nil {
		t.Error("WriteTo should return the error")
	}
	f.Reset()
	if f.Err() != nil {
		t.Error("Reset should clear the error")
	}

	f.SetEncoder(func(s string) string { return "<" + s + ">" })
	f.WriteString(1, "text")
	f.WriteFloat(40, math.Pi)
	out = []byte(f.Output())[len(BinarySentinel):]
	if !bytes.HasPrefix(out[2:], []byte("<text>\x00")) {
		t.Errorf("encoder should be applied to strings: %q", out)
	}
}
