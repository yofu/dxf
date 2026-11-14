package entity

import (
	"github.com/yofu/dxf/format"
)

// Viewport represents VIEWPORT Entity: a view of model space placed in
// paper space. The viewport with ID 1 is the paper space view itself.
type Viewport struct {
	*entity
	Center        []float64 // 10, 20, 30: center in paper space
	Width         float64   // 40: width in paper space
	Height        float64   // 41: height in paper space
	Status        int       // 68: 0 off, -1 on but off screen, >0 on (stacking order)
	ID            int       // 69
	ViewCenter    []float64 // 12, 22: view center in model space (DCS)
	SnapBase      []float64 // 13, 23
	SnapSpacing   []float64 // 14, 24
	GridSpacing   []float64 // 15, 25
	ViewDirection []float64 // 16, 26, 36
	ViewTarget    []float64 // 17, 27, 37
	LensLength    float64   // 42
	FrontClip     float64   // 43
	BackClip      float64   // 44
	ViewHeight    float64   // 45: height of the view in model space
	SnapAngle     float64   // 50
	TwistAngle    float64   // 51
	CircleZoom    int       // 72
	Flags         int       // 90: viewport status flags
	PlotStyle     string    // 1: plot style sheet name
	RenderMode    int       // 281
	UCSPerView    int       // 71
	UCSIcon       int       // 74
	UCSOrigin     []float64 // 110, 120, 130
	UCSXAxis      []float64 // 111, 121, 131
	UCSYAxis      []float64 // 112, 122, 132
	OrthoType     int       // 79
	Elevation     float64   // 146
	// References holds references by handle as read: frozen layers (331),
	// clipping boundary (340) and UCS (345, 346).
	References [][2]string
}

// IsEntity is for Entity interface.
func (v *Viewport) IsEntity() bool {
	return true
}

// NewViewport creates a new Viewport in paper space.
func NewViewport() *Viewport {
	v := &Viewport{
		entity:        NewEntity(VIEWPORT),
		Center:        []float64{0, 0, 0},
		Width:         1,
		Height:        1,
		Status:        1,
		ID:            1,
		ViewCenter:    []float64{0, 0},
		SnapBase:      []float64{0, 0},
		SnapSpacing:   []float64{10, 10},
		GridSpacing:   []float64{10, 10},
		ViewDirection: []float64{0, 0, 1},
		ViewTarget:    []float64{0, 0, 0},
		LensLength:    50,
		ViewHeight:    1,
		CircleZoom:    100,
		Flags:         32864,
		UCSIcon:       3,
		UCSOrigin:     []float64{0, 0, 0},
		UCSXAxis:      []float64{1, 0, 0},
		UCSYAxis:      []float64{0, 1, 0},
	}
	v.paperSpace = true
	return v
}

// Scale returns the scale of the view: paper space units per model space unit.
func (v *Viewport) Scale() float64 {
	if v.ViewHeight == 0 {
		return 0
	}
	return v.Height / v.ViewHeight
}

// Format writes data to formatter.
func (v *Viewport) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbViewport")
	writePoint(f, 10, v.Center)
	f.WriteFloat(40, v.Width)
	f.WriteFloat(41, v.Height)
	f.WriteInt(68, v.Status)
	f.WriteInt(69, v.ID)
	for i, p := range [][]float64{v.ViewCenter, v.SnapBase, v.SnapSpacing, v.GridSpacing} {
		f.WriteFloat(12+i, p[0])
		f.WriteFloat(22+i, p[1])
	}
	writePoint(f, 16, v.ViewDirection)
	writePoint(f, 17, v.ViewTarget)
	f.WriteFloat(42, v.LensLength)
	f.WriteFloat(43, v.FrontClip)
	f.WriteFloat(44, v.BackClip)
	f.WriteFloat(45, v.ViewHeight)
	f.WriteFloat(50, v.SnapAngle)
	f.WriteFloat(51, v.TwistAngle)
	f.WriteInt(72, v.CircleZoom)
	v.writeReferences(f, "331")
	f.WriteInt(90, v.Flags)
	v.writeReferences(f, "340")
	f.WriteString(1, v.PlotStyle)
	f.WriteInt(281, v.RenderMode)
	f.WriteInt(71, v.UCSPerView)
	f.WriteInt(74, v.UCSIcon)
	writePoint(f, 110, v.UCSOrigin)
	writePoint(f, 111, v.UCSXAxis)
	writePoint(f, 112, v.UCSYAxis)
	v.writeReferences(f, "345", "346")
	f.WriteInt(79, v.OrthoType)
	f.WriteFloat(146, v.Elevation)
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

// BBox returns the bounding box of the viewport in paper space.
func (v *Viewport) BBox() ([]float64, []float64) {
	return []float64{v.Center[0] - v.Width/2, v.Center[1] - v.Height/2, v.Center[2]},
		[]float64{v.Center[0] + v.Width/2, v.Center[1] + v.Height/2, v.Center[2]}
}

func (v *Viewport) writeReferences(f format.Formatter, codes ...string) {
	for _, r := range v.References {
		for _, c := range codes {
			if r[0] == c {
				writeRaw(f, [][2]string{r})
			}
		}
	}
}
