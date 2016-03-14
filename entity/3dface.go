package entity

import (
	"github.com/yofu/dxf/format"
)

// ThreeDFace represents 3DFACE Entity.
type ThreeDFace struct {
	*entity
	Points [][]float64
	Flag   int // 70
}

// IsEntity is for Entity interface.
func (f *ThreeDFace) IsEntity() bool {
	return true
}

// New3DFace creates a new ThreeDFace.
func New3DFace() *ThreeDFace {
	f := &ThreeDFace{
		entity: NewEntity(THREEDFACE),
		Points: [][]float64{[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
		},
		Flag: 0,
	}
	return f
}

// Format writes data to formatter.
func (f *ThreeDFace) Format(fm format.Formatter) {
	f.entity.Format(fm)
	fm.WriteString(100, "AcDbFace")
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			fm.WriteFloat((j+1)*10+i, f.Points[i][j])
		}
	}
	if f.Flag != 0 {
		fm.WriteInt(70, f.Flag)
	}
}

// String outputs data using default formatter.
func (f *ThreeDFace) String() string {
	fm := format.NewASCII()
	return f.FormatString(fm)
}

// FormatString outputs data using given formatter.
func (f *ThreeDFace) FormatString(fm format.Formatter) string {
	f.Format(fm)
	return fm.Output()
}

func (f *ThreeDFace) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	for _, p := range f.Points {
		for i := 0; i < 3; i++ {
			if p[i] < mins[i] {
				mins[i] = p[i]
			}
			if p[i] > maxs[i] {
				maxs[i] = p[i]
			}
		}
	}
	return mins, maxs
}
