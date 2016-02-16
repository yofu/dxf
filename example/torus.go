package main

import (
	"github.com/yofu/dxf"
	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/table"
	"log"
	"math"
)

func main() {
	d := dxf.NewDrawing()
	d.AddLayer("Toroidal", dxf.DefaultColor, dxf.DefaultLineType, true)
	d.AddLayer("Poloidal", color.Red, table.LT_HIDDEN, true)
	z := 0.0
	r1 := 200.0
	r2 := 500.0
	ndiv := 16
	dtheta := 2.0 * math.Pi / float64(ndiv)
	theta := 0.0
	for i := 0; i < ndiv; i++ {
		d.ChangeLayer("Toroidal")
		d.Circle(0.0, 0.0, z+r1*math.Cos(theta), r2-r1*math.Sin(theta))
		d.ChangeLayer("Poloidal")
		c, _ := d.Circle(r2*math.Cos(theta), r2*math.Sin(theta), 0.0, r1)
		dxf.SetExtrusion(c, []float64{-1.0 * math.Sin(theta), math.Cos(theta), 0.0})
		theta += dtheta
	}
	err := d.SaveAs("torus.dxf")
	if err != nil {
		log.Fatal(err)
	}
}
