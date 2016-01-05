package header

import (
	"github.com/yofu/dxf/format"
)

type Header struct {
	Version  string
	InsBase  []float64
	ExtMin   []float64
	ExtMax   []float64
	handseed int
}

func New() *Header {
	h := new(Header)
	h.Version = "AC1015"
	h.InsBase = make([]float64, 3)
	h.ExtMin = make([]float64, 3)
	h.ExtMax = make([]float64, 3)
	return h
}

func (h *Header) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	f.WriteString(9, "$ACADVER")
	f.WriteString(1, h.Version)
	f.WriteString(9, "$INSBASE")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.InsBase[i])
	}
	f.WriteString(9, "$EXTMIN")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMin[i])
	}
	f.WriteString(9, "$EXTMAX")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMax[i])
	}
	f.WriteString(9, "$HANDSEED")
	f.WriteHex(5, h.handseed)
	f.WriteString(0, "ENDSEC")
}

func (h *Header) SetHandle(v *int) {
	h.handseed = *v
}
