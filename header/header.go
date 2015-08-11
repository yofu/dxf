package header

import (
	"bytes"
	"fmt"
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

func (h *Header) WriteTo(b *bytes.Buffer) error {
	b.WriteString("0\nSECTION\n2\nHEADER\n")
	b.WriteString(fmt.Sprintf("9\n$ACADVER\n1\n%s\n", h.version))
	b.WriteString("9\n$INSBASE\n")
	for i := 0; i < 3; i++ {
		b.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, h.insbase[i]))
	}
	b.WriteString("9\n$EXTMIN\n")
	for i := 0; i < 3; i++ {
		b.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, h.extmin[i]))
	}
	b.WriteString("9\n$EXTMAX\n")
	for i := 0; i < 3; i++ {
		b.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, h.extmax[i]))
	}
	b.WriteString(fmt.Sprintf("9\n$HANDSEED\n5\n%x\n", h.handseed))
	b.WriteString("0\nENDSEC\n")
	return nil
}

func (h *Header) SetHandle(v *int) {
	h.handseed = *v
}
