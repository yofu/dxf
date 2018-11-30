package entity

import (
	"github.com/yofu/dxf/format"
)

// Spline represents LINE Entity.
type Spline struct {
	*entity
	Normal    []float64   // 210, 220, 230
	Flag      int         // 70
	Degree    int         // 71
	Knots     []float64   // 72, 40
	Controls  [][]float64 // 73, 10, 20, 30
	Fits      [][]float64 // 74, 11, 21, 31
	Tolerance []float64   // 42, 43, 44
}

// IsEntity is for Entity interface.
func (l *Spline) IsEntity() bool {
	return true
}

// NewSpline creates a new Spline.
func NewSpline() *Spline {
	l := &Spline{
		entity:    NewEntity(SPLINE),
		Normal:    []float64{0.0, 0.0, 1.0},
		Flag:      1064,
		Degree:    3,
		Knots:     nil,
		Controls:  nil,
		Fits:      nil,
		Tolerance: []float64{0.000000001, 0.0000000001, 0.0000000001},
	}
	return l
}

// Format writes data to formatter.
func (s *Spline) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbSpline")
	for i := 0; i < 3; i++ {
		f.WriteFloat(210+i*10, s.Normal[i])
	}
	f.WriteInt(70, s.Flag)
	f.WriteInt(71, s.Degree)
	f.WriteInt(72, len(s.Knots))
	f.WriteInt(73, len(s.Controls))
	f.WriteInt(74, len(s.Fits))
	for i := 0; i < 3; i++ {
		f.WriteFloat(42+i, s.Tolerance[i])
	}
	for _, k := range s.Knots {
		f.WriteFloat(40, k)
	}
	for _, c := range s.Controls {
		for i := 0; i < 3; i++ {
			f.WriteFloat((i+1)*10, c[i])
		}
	}
	for _, ft := range s.Fits {
		for i := 0; i < 3; i++ {
			f.WriteFloat((i+1)*10+1, ft[i])
		}
	}
}

// String outputs data using default formatter.
func (s *Spline) String() string {
	f := format.NewASCII()
	return s.FormatString(f)
}

// FormatString outputs data using given formatter.
func (s *Spline) FormatString(f format.Formatter) string {
	s.Format(f)
	return f.Output()
}

// func (s *Spline) BBox() ([]float64, []float64) {
// 	mins := make([]float64, 3)
// 	maxs := make([]float64, 3)
// 	for i := 0; i < 3; i++ {
// 		if l.Start[i] <= l.End[i] {
// 			mins[i] = l.Start[i]
// 			maxs[i] = l.End[i]
// 		} else {
// 			mins[i] = l.End[i]
// 			maxs[i] = l.Start[i]
// 		}
// 	}
// 	return mins, maxs
// }
