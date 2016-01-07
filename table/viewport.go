package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Viewport represents VPORT SymbolTable.
type Viewport struct {
	handle        int
	owner         handle.Handler
	name          string // 2
	LowerLeft     []float64
	UpperRight    []float64
	ViewCenter    []float64
	SnapBase      []float64
	SnapSpacing   []float64
	GridSpacing   []float64
	ViewDirection []float64
	ViewTarget    []float64
	Height        float64
	AspectRatio   float64
	LensLength    float64
	FrontClip     float64
	BackClip      float64
	SnapAngle     float64
	TwistAngle    float64
}

// NewViewport creates a new Viewport.
func NewViewport(name string) *Viewport {
	v := &Viewport{
		name:          name,
		LowerLeft:     []float64{0.0, 0.0},
		UpperRight:    []float64{0.0, 0.0},
		ViewCenter:    []float64{0.0, 0.0},
		SnapBase:      []float64{0.0, 0.0},
		SnapSpacing:   []float64{0.0, 0.0},
		GridSpacing:   []float64{0.0, 0.0},
		ViewDirection: []float64{0.0, 0.0, 0.0},
		ViewTarget:    []float64{0.0, 0.0, 0.0},
		Height:        400.0,
		AspectRatio:   1.0,
		LensLength:    50.0,
		FrontClip:     0.0,
		BackClip:      0.0,
		SnapAngle:     0.0,
		TwistAngle:    0.0,
	}
	return v
}

// IsSymbolTable is for SymbolTable interface.
func (v *Viewport) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (v *Viewport) Format(f format.Formatter) {
	f.WriteString(0, "VPORT")
	f.WriteHex(5, v.handle)
	if v.owner != nil {
		f.WriteHex(330, v.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbViewportTableRecord")
	f.WriteString(2, v.name)
	f.WriteInt(70, 0)
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10, v.LowerLeft[i])
	}
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10+1, v.UpperRight[i])
	}
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10+2, v.ViewCenter[i])
	}
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10+3, v.SnapBase[i])
	}
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10+4, v.SnapSpacing[i])
	}
	for i := 0; i < 2; i++ {
		f.WriteFloat((i+1)*10+5, v.GridSpacing[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10+6, v.ViewDirection[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10+7, v.ViewTarget[i])
	}
	f.WriteFloat(40, v.Height)
	f.WriteFloat(41, v.AspectRatio)
	f.WriteFloat(42, v.LensLength)
	f.WriteFloat(43, v.FrontClip)
	f.WriteFloat(44, v.BackClip)
	f.WriteFloat(50, v.SnapAngle)
	f.WriteFloat(51, v.TwistAngle)
}

// String outputs data using default formatter.
func (v *Viewport) String() string {
	f := format.NewASCII()
	return v.FormatString(f)
}

// FormatString outputs data using given formatter.
func (v *Viewport) FormatString(f format.Formatter) string {
	v.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (v *Viewport) Handle() int {
	return v.handle
}

// SetHandle sets a handle.
func (v *Viewport) SetHandle(h *int) {
	v.handle = *h
	(*h)++
}

// SetOwner sets an owner.
func (v *Viewport) SetOwner(h handle.Handler) {
	v.owner = h
}

// Name returns a name of VPORT (code 2).
func (v *Viewport) Name() string {
	return v.name
}

// SetName sets a name to VPORT (code 2).
func (v *Viewport) SetName(name string) {
	v.name = name
}
