package header

import (
	"github.com/yofu/dxf/format"
)

type Header struct {
	version  string
	insbase  []float64
	extmin   []float64
	extmax   []float64
	handseed int
}

func New() *Header {
	h := new(Header)
	h.version = "AC1015"
	h.insbase = make([]float64, 3)
	h.extmin = make([]float64, 3)
	h.extmax = make([]float64, 3)
	return h
}

func (h *Header) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	f.WriteString(9, "$ACADVER")
	f.WriteString(1, h.version)
	f.WriteString(9, "$INSBASE")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.insbase[i])
	}
	f.WriteString(9, "$EXTMIN")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.extmin[i])
	}
	f.WriteString(9, "$EXTMAX")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.extmax[i])
	}
	f.WriteString(9, "$HANDSEED")
	f.WriteHex(5, h.handseed)
	f.WriteString(0, "ENDSEC")
}

func (h *Header) SetHandle(v *int) {
	h.handseed = *v
}
