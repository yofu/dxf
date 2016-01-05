package entity

import (
	"github.com/yofu/dxf/format"
)

type ThreeDFace struct {
	*entity
	Points [][]float64
	Flag   int // 70
}

func (f *ThreeDFace) IsEntity() bool {
	return true
}

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

func (f *ThreeDFace) Format(fm *format.Formatter) {
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

func (f *ThreeDFace) String() string {
	fm := format.New()
	return f.FormatString(fm)
}

func (f *ThreeDFace) FormatString(fm *format.Formatter) string {
	f.Format(fm)
	return fm.Output()
}
