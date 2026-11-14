package entity

import (
	"math"

	"github.com/yofu/dxf/format"
)

// LwPolyline represents LWPOLYLINE Entity.
type LwPolyline struct {
	*entity
	Num           int         // 90
	Closed        bool        // 70 & 1
	Plinegen      bool        // 70 & 128: linetype generated continuously around vertices
	ConstantWidth float64     // 43 (used when Widths are all 0)
	Elevation     float64     // 38
	Thickness     float64     // 39
	Vertices      [][]float64 // 10, 20
	Widths        [][]float64 // 40, 41: starting and ending width of each segment
	Bulges        []float64   // 42
	Extrusion     []float64   // 210, 220, 230
}

// IsEntity is for Entity interface.
func (l *LwPolyline) IsEntity() bool {
	return true
}

// NewLwPolyline creates a new LwPolyline.
func NewLwPolyline(size int) *LwPolyline {
	vs := make([][]float64, size)
	for i := 0; i < size; i++ {
		vs[i] = make([]float64, 2)
	}
	l := &LwPolyline{
		entity:   NewEntity(LWPOLYLINE),
		Num:      size,
		Closed:   false,
		Vertices: vs,
		Bulges:   make([]float64, size),
	}
	l.Widths = make([][]float64, size)
	for i := range l.Widths {
		l.Widths[i] = []float64{0.0, 0.0}
	}
	l.Extrusion = []float64{0.0, 0.0, 1.0}
	return l
}

// SetWidth sets the starting and ending width of the segment from vertex i.
func (l *LwPolyline) SetWidth(i int, start, end float64) {
	for len(l.Widths) <= i {
		l.Widths = append(l.Widths, []float64{0.0, 0.0})
	}
	l.Widths[i] = []float64{start, end}
}

// width returns the widths of segment i (0 if not set).
func (l *LwPolyline) width(i int) (float64, float64) {
	if i < len(l.Widths) && len(l.Widths[i]) >= 2 {
		return l.Widths[i][0], l.Widths[i][1]
	}
	return 0, 0
}

// Format writes data to formatter.
func (l *LwPolyline) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbPolyline")
	f.WriteInt(90, l.Num)
	flag := 0
	if l.Closed {
		flag |= 1
	}
	if l.Plinegen {
		flag |= 128
	}
	f.WriteInt(70, flag)
	hasWidth := false
	for i := 0; i < l.Num; i++ {
		if s, e := l.width(i); s != 0 || e != 0 {
			hasWidth = true
		}
	}
	if !hasWidth && l.ConstantWidth != 0 {
		f.WriteFloat(43, l.ConstantWidth)
	}
	if l.Elevation != 0 {
		f.WriteFloat(38, l.Elevation)
	}
	writeThickness(f, l.Thickness)
	for i := 0; i < l.Num; i++ {
		for j := 0; j < 2; j++ {
			f.WriteFloat((j+1)*10, l.Vertices[i][j])
		}
		if hasWidth {
			s, e := l.width(i)
			f.WriteFloat(40, s)
			f.WriteFloat(41, e)
		}
		if i < len(l.Bulges) && l.Bulges[i] != 0 {
			f.WriteFloat(42, l.Bulges[i])
		}
	}
	if len(l.Extrusion) == 3 && (l.Extrusion[0] != 0 || l.Extrusion[1] != 0 || l.Extrusion[2] != 1) {
		for i := 0; i < 3; i++ {
			f.WriteFloat(200+(i+1)*10, l.Extrusion[i])
		}
	}
}

// String outputs data using default formatter.
func (l *LwPolyline) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *LwPolyline) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// Close closes LwPolyline.
func (l *LwPolyline) Close() {
	l.Closed = true
}

func (l *LwPolyline) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for i := 0; i < len(mins) && len(l.Vertices) > 0; i++ {
		if i < len(l.Vertices[0]) {
			mins[i] = l.Vertices[0][i]
			maxs[i] = l.Vertices[0][i]
		}
	}
	for _, p := range l.Vertices {
		for i := 0; i < len(p) && i < 3; i++ {
			if p[i] < mins[i] {
				mins[i] = p[i]
			}
			if p[i] > maxs[i] {
				maxs[i] = p[i]
			}
		}
	}
	// vertices are 2D (OCS); z is the elevation
	mins[2], maxs[2] = l.Elevation, l.Elevation
	// widths extend the outline by half of the largest width
	w := l.ConstantWidth
	for i := 0; i < l.Num; i++ {
		s, e := l.width(i)
		w = math.Max(w, math.Max(s, e))
	}
	for i := 0; i < 2; i++ {
		mins[i] -= w / 2
		maxs[i] += w / 2
	}
	return extrude(mins, maxs, l.Thickness)
}

// Bulge computes the center and the radius of the arc to reach the vertex.
// The bulge of a segment is stored on its start vertex (Bulges[index-1]).
func (l LwPolyline) Bulge(index int) ([]float64, float64) {
	if index > len(l.Vertices)-1 || index < 1 {
		return nil, 0
	}

	return bulge(l.Bulges[index-1], l.Vertices[index-1], l.Vertices[index])
}
