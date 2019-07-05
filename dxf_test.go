package dxf

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yofu/dxf/color"
	"github.com/yofu/dxf/insunit"
	"github.com/yofu/dxf/table"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
)

// TOLERANCE is the epsilon value used in comparing floats.
const TOLERANCE = 0.000001

// cmpF64 compares two floats to see if they are within the given tolerance.
func cmpF64(f1, f2 float64) bool {
	if math.IsInf(f1, 1) {
		return math.IsInf(f2, 1)
	}
	if math.IsInf(f2, 1) {
		return math.IsInf(f1, 1)
	}
	if math.IsInf(f1, -1) {
		return math.IsInf(f2, -1)
	}
	if math.IsInf(f2, -1) {
		return math.IsInf(f1, -1)
	}
	return math.Abs(f1-f2) < TOLERANCE
}

func checkEntities(t *testing.T, expected, got entity.Entities) bool {
	if len(expected) != len(got) {
		t.Errorf("number of entities, expected %v got %v", len(expected), len(got))
		return false
	}
	for i, ee := range expected {
		switch et := ee.(type) {
		case *entity.Point:
			p, ok := got[i].(*entity.Point)
			if !ok {
				t.Errorf("type for %v, expected %T got %T", i, et, got[i])
				return false
			}
			if len(et.Coord) != len(p.Coord) {
				t.Errorf("point %v : coord len, expected %v got %v", i, len(et.Coord), len(p.Coord))
				return false
			}
			for j, crd := range et.Coord {
				if !cmpF64(crd, p.Coord[j]) {
					t.Errorf("point %v : coord %v, expected %v got %v", i, j, crd, p.Coord[j])
					t.Logf("ePoint %v, Point %v", et.Coord, p.Coord)
					return false
				}
			}
		}
	}
	return true
}

func hashBytes(b []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(b))
}
func hashFile(filename string) (string, error) {
	tfile := filepath.Join("testdata", filename)
	f, err := os.Open(tfile)
	if err != nil {
		return "", err
	}
	text, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return "", err
	}
	return hashBytes(text), nil
}

func TestFromStringData(t *testing.T) {
	type tcase struct {
		filename         string
		ExpectedEntities entity.Entities
		fn               func(*testing.T, *drawing.Drawing, tcase)
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.filename, func(t *testing.T) {
			tfile := filepath.Join("testdata", tc.filename)

			data, err := ioutil.ReadFile(tfile)
			if err != nil {
				t.Errorf("file, could not open file %v : %v", tfile, err)
			}
			d, err := FromStringData(string(data))
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}

			if len(tc.ExpectedEntities) != 0 && !checkEntities(t, tc.ExpectedEntities, d.Entities()) {
				return
			}

			if tc.fn != nil {
				tc.fn(t, d, tc)
			}
		}
	}

	tests := []tcase{
		{
			filename: "point.dxf",
			ExpectedEntities: entity.Entities{
				entity.NewPoint(),
				entity.NewPoint(100.0, 100.0, 0.0),
				entity.NewPoint(200.0, 100.0, 0.0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(fn(tc))
	}
}

func TestFromFile(t *testing.T) {
	type tcase struct {
		filename         string
		ExpectedEntities entity.Entities
		fn               func(*testing.T, *drawing.Drawing, tcase)
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.filename, func(t *testing.T) {
			tfile := filepath.Join("testdata", tc.filename)
			d, err := FromFile(tfile)
			if err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}

			if len(tc.ExpectedEntities) != 0 && !checkEntities(t, tc.ExpectedEntities, d.Entities()) {
				return
			}

			if tc.fn != nil {
				tc.fn(t, d, tc)
			}
		}
	}

	tests := []tcase{
		{
			filename: "point.dxf",
			ExpectedEntities: entity.Entities{
				entity.NewPoint(),
				entity.NewPoint(100.0, 100.0, 0.0),
				entity.NewPoint(200.0, 100.0, 0.0),
			},
		},
		{
			filename: "mypoint.dxf",
			ExpectedEntities: entity.Entities{
				entity.NewPoint(),
				entity.NewPoint(100.0, 100.0, 0.0),
				entity.NewPoint(100.0, 200.0, 0.0),
			},
		},
		{
			filename: "mypoint_with_extent.dxf",
			ExpectedEntities: entity.Entities{
				entity.NewPoint(),
				entity.NewPoint(100.0, 100.0, 0.0),
				entity.NewPoint(100.0, 200.0, 0.0),
			},
		},
	}
	for _, tc := range tests {
		t.Run(fn(tc))
	}
}

func TestNewDrawing(t *testing.T) {
	type tcase struct {
		filename string
		draw     func(d *drawing.Drawing)
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.filename, func(t *testing.T) {
			// calculate the sha256
			fileHash, err := hashFile(tc.filename)
			if err != nil {
				t.Errorf("hash of file(%v) error, expected nil got %v", tc.filename, err)
				return
			}
			d := drawing.New()
			tc.draw(d)
			var buff bytes.Buffer
			_, err = io.Copy(&buff, d)
			if err != nil {
				t.Errorf("copy of buffer, expected nil got %v", err)
			}
			buffHash := hashBytes(buff.Bytes())
			if fileHash != buffHash {
				t.Errorf("hash, expected %v got %v", fileHash, buffHash)
				outputfn := filepath.Join("testdata", buffHash+"_"+tc.filename)
				of, err := os.Create(outputfn)
				if err != nil {
					t.Logf("could not create debug file: %v", outputfn)
					return
				}
				io.Copy(of, &buff)
				of.Close()
				t.Logf("wrote out file to: %v", outputfn)
			}
		}
	}
	tests := []tcase{
		{
			filename: "mypoint.dxf",
			draw: func(d *drawing.Drawing) {
				d.Point(0.0, 0.0, 0.0)
				d.Point(100.0, 100.0, 0.0)
				d.Point(100.0, 200.0, 0.0)
			},
		},
		{
			filename: "mypoint_with_extent.dxf",
			draw: func(d *drawing.Drawing) {
				d.Point(0.0, 0.0, 0.0)
				d.Point(100.0, 100.0, 0.0)
				d.Point(100.0, 200.0, 0.0)
				d.SetExt()
			},
		},
		{
			filename: "mypoint_with_units.dxf",
			draw: func(d *drawing.Drawing) {
				d.Point(0.0, 0.0, 0.0)
				d.Point(100.0, 100.0, 0.0)
				d.Point(100.0, 200.0, 0.0)
				d.Header().LtScale = 1
				d.Header().InsUnit = insunit.Inches
				d.Header().InsLUnit = insunit.Architectural
				d.SetExt()
			},
		},
		{
			filename: "torus.dxf",
			draw: func(d *drawing.Drawing) {
				d.Header().LtScale = 100.0
				d.AddLayer("Toroidal", DefaultColor, DefaultLineType, true)
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
					SetExtrusion(c, []float64{-1.0 * math.Sin(theta), math.Cos(theta), 0.0})
					theta += dtheta
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(fn(tc))
	}

}
