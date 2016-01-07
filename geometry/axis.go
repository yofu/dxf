// Geometrical Functions
package geometry

import (
	"errors"
	"math"
)

// ArbitraryAxis generates according X, Y axis from given Z axis using Arbitrary Axis Algorithm.
// Z axis is assumed to be unit-length vector.
// http://www.autodesk.com/techpubs/autocad/acad2000/dxf/arbitrary_axis_algorithm_dxf_ab.htm
func ArbitraryAxis(d []float64) ([]float64, []float64, error) {
	if len(d) < 3 {
		return nil, nil, errors.New("not enough length")
	}
	thres := 1.0 / 64.0
	ax := make([]float64, 3)
	ay := make([]float64, 3)
	if math.Abs(d[0]) < thres && math.Abs(d[1]) < thres {
		sum := math.Sqrt(d[1]*d[1] + d[2]*d[2])
		ax[0] = d[2] / sum
		ax[1] = 0.0
		ax[2] = -d[1] / sum
	} else {
		sum := math.Sqrt(d[0]*d[0] + d[1]*d[1])
		ax[0] = -d[1] / sum
		ax[1] = d[0] / sum
		ax[2] = 0.0
	}
	ay[0] = d[1]*ax[2] - d[2]*ax[1]
	ay[1] = d[2]*ax[0] - d[0]*ax[2]
	ay[2] = d[0]*ax[1] - d[1]*ax[0]
	return ax, ay, nil
}
