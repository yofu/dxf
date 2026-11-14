// HEADER section
package header

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/insunit"
)

// Measurement systems ($MEASUREMENT).
const (
	Imperial = 0
	Metric   = 1
)

// Header contains information written in HEADER section.
type Header struct {
	Version  string
	CodePage string // $DWGCODEPAGE: code page of text in the file (e.g. "ANSI_932" for Japanese)
	InsBase  []float64
	InsUnit  insunit.Unit
	InsLUnit insunit.Type
	ExtMin   []float64
	ExtMax   []float64
	LimMin   []float64 // $LIMMIN: lower-left corner of the drawing limits (model space)
	LimMax   []float64 // $LIMMAX: upper-right corner of the drawing limits
	LtScale  float64

	// current properties for new entities
	CLayer    string  // $CLAYER: current layer (kept in sync with Drawing.CurrentLayer)
	CEColor   int     // $CECOLOR: current color (0 BYBLOCK, 1-255, 256 BYLAYER)
	CELType   string  // $CELTYPE: current linetype ("ByLayer", "ByBlock" or a name)
	CELWeight int     // $CELWEIGHT: current lineweight (-1 BYLAYER, -2 BYBLOCK, -3 DEFAULT, 0-211)
	CELTScale float64 // $CELTSCALE: current linetype scale

	TextStyle string  // $TEXTSTYLE: current text style (kept in sync with Drawing.CurrentStyle)
	TextSize  float64 // $TEXTSIZE: default text height
	DimStyle  string  // $DIMSTYLE: current dimension style

	Measurement int     // $MEASUREMENT: Imperial (0) or Metric (1)
	LUPrec      int     // $LUPREC: decimal places of linear units
	AUnits      int     // $AUNITS: angle units (0 decimal degrees, 1 deg/min/sec, 2 grads, 3 radians, 4 surveyor's)
	AUPrec      int     // $AUPREC: decimal places of angular units
	AngBase     float64 // $ANGBASE: angle 0 direction (degrees)
	AngDir      int     // $ANGDIR: 0 counterclockwise, 1 clockwise
	PdMode      int     // $PDMODE: point display mode
	PdSize      float64 // $PDSIZE: point display size

	handseed int
}

// New creates a new Header.
func New() *Header {
	h := new(Header)
	h.Version = "AC1015"
	h.CodePage = "ANSI_1252"
	h.InsBase = make([]float64, 3)
	h.ExtMin = make([]float64, 3)
	h.ExtMax = make([]float64, 3)
	h.LimMin = []float64{0.0, 0.0}
	h.LimMax = []float64{420.0, 297.0}
	h.LtScale = 1.0
	h.CLayer = "0"
	h.CEColor = 256
	h.CELType = "ByLayer"
	h.CELWeight = -1
	h.CELTScale = 1.0
	h.TextStyle = "Standard"
	h.TextSize = 2.5
	h.DimStyle = "Standard"
	h.Measurement = Metric
	h.LUPrec = 4
	return h
}

func writeVar(f format.Formatter, name string) {
	f.WriteString(9, name)
}

// Format writes HEADER information to formatter.
func (h *Header) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "HEADER")
	writeVar(f, "$ACADVER")
	f.WriteString(1, h.Version)
	if h.CodePage != "" {
		writeVar(f, "$DWGCODEPAGE")
		f.WriteString(3, h.CodePage)
	}
	writeVar(f, "$INSBASE")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.InsBase[i])
	}
	writeVar(f, "$INSUNITS")
	h.InsUnit.Format(f)
	writeVar(f, "$LUNITS")
	h.InsLUnit.Format(f)
	writeVar(f, "$LUPREC")
	f.WriteInt(70, h.LUPrec)

	writeVar(f, "$EXTMIN")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMin[i])
	}
	writeVar(f, "$EXTMAX")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, h.ExtMax[i])
	}
	writeVar(f, "$LIMMIN")
	f.WriteFloat(10, h.LimMin[0])
	f.WriteFloat(20, h.LimMin[1])
	writeVar(f, "$LIMMAX")
	f.WriteFloat(10, h.LimMax[0])
	f.WriteFloat(20, h.LimMax[1])
	writeVar(f, "$LTSCALE")
	f.WriteFloat(40, h.LtScale)
	writeVar(f, "$TEXTSIZE")
	f.WriteFloat(40, h.TextSize)
	writeVar(f, "$CLAYER")
	f.WriteString(8, h.CLayer)
	writeVar(f, "$CELTYPE")
	f.WriteString(6, h.CELType)
	writeVar(f, "$CECOLOR")
	f.WriteInt(62, h.CEColor)
	writeVar(f, "$CELTSCALE")
	f.WriteFloat(40, h.CELTScale)
	writeVar(f, "$DIMSTYLE")
	f.WriteString(2, h.DimStyle)
	writeVar(f, "$AUNITS")
	f.WriteInt(70, h.AUnits)
	writeVar(f, "$AUPREC")
	f.WriteInt(70, h.AUPrec)
	writeVar(f, "$PDMODE")
	f.WriteInt(70, h.PdMode)
	writeVar(f, "$PDSIZE")
	f.WriteFloat(40, h.PdSize)
	writeVar(f, "$TEXTSTYLE")
	f.WriteString(7, h.TextStyle)
	writeVar(f, "$ANGBASE")
	f.WriteFloat(50, h.AngBase)
	writeVar(f, "$ANGDIR")
	f.WriteInt(70, h.AngDir)
	writeVar(f, "$HANDSEED")
	f.WriteHex(5, h.handseed)
	writeVar(f, "$MEASUREMENT")
	f.WriteInt(70, h.Measurement)
	writeVar(f, "$CELWEIGHT")
	f.WriteInt(370, h.CELWeight)
	f.WriteString(0, "ENDSEC")
}

// SetHandle sets $HANDSEED.
func (h *Header) SetHandle(v *int) {
	h.handseed = *v
}
