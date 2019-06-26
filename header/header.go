// HEADER section
package header

import (
	"github.com/gdey/dxf/format"
	"github.com/gdey/dxf/insunit"
)

// Header contains information written in HEADER section.
type Header struct {
	Version  string
	InsBase  []float64
	InsUnit  insunit.Unit
	InsLUnit insunit.Type
	ExtMin   []float64
	ExtMax   []float64
	LtScale  float64
	handseed int
}

// New creates a new Header.
func New() *Header {
	h := new(Header)
	h.Version = "AC1015"
	h.InsBase = make([]float64, 3)
	h.ExtMin = make([]float64, 3)
	h.ExtMax = make([]float64, 3)
	h.LtScale = 1.0
	return h
}

// WriteTo writes HEADER information to formatter.
func (h *Header) WriteTo(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	f.WriteString(9, "$ACADVER")
	f.WriteString(1, h.Version)
	f.WriteString(9, "$INSBASE")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.InsBase[i])
	}
	f.WriteString(9, "$INSUNITS")
	h.InsUnit.Format(f)
	f.WriteString(9, "$LUNITS")
	h.InsLUnit.Format(f)

	f.WriteString(9, "$EXTMIN")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMin[i])
	}
	f.WriteString(9, "$EXTMAX")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMax[i])
	}
	f.WriteString(9, "$LTSCALE")
	f.WriteFloat(40, h.LtScale)
	f.WriteString(9, "$HANDSEED")
	f.WriteHex(5, h.handseed)
	f.WriteString(0, "ENDSEC")
}

// SetHandle sets $HANDSEED.
func (h *Header) SetHandle(v *int) {
	h.handseed = *v
}
