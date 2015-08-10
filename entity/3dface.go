package entity

import (
	"bytes"
	"fmt"
)

type ThreeDFace struct {
	*entity
	Points    [][]float64
	Flag      int // 70
}

func (f *ThreeDFace) IsEntity() bool {
	return true
}

func New3DFace() *ThreeDFace {
	f := &ThreeDFace{
		NewEntity(THREEDFACE),
		[][]float64{[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
			[]float64{0.0, 0.0, 0.0},
		},
		0,
	}
	return f
}

func (f *ThreeDFace) String() string {
	var otp bytes.Buffer
	otp.WriteString(f.entity.String())
	otp.WriteString("100\nAcDbFace\n")
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			otp.WriteString(fmt.Sprintf("%d\n%f\n", (j+1)*10+i, f.Points[i][j]))
		}
	}
	if f.Flag != 0 {
		otp.WriteString(fmt.Sprintf("70\n%d\n", f.Flag))
	}
	return otp.String()
}

