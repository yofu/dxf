package dxf

import (
	"github.com/yofu/dxf/geometry"
)

type Extruder interface { // 210 220 230
	CurrentDirection() []float64
	SetDirection([]float64)
	CurrentCoord() []float64
	SetCoord([]float64)
}

func SetExtrusion(e Extruder, d []float64) {
	dx, dy, err := geometry.ArbitraryAxis(d)
	if err != nil {
		return
	}
	b := e.CurrentDirection()
	e.SetDirection(d)
	n := e.CurrentCoord()
	bx, by, _ := geometry.ArbitraryAxis(b)
	before := [][]float64{bx, by, b}
	after := [][]float64{dx, dy, d}
	newcoord := make([]float64, 3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				newcoord[i] += n[j] * before[j][k] * after[i][k]
			}
		}
	}
	e.SetCoord(newcoord)
}
