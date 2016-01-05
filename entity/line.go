package entity

import (
	"fmt"
	"github.com/yofu/dxf/format"
	"strconv"
)

type Line struct {
	*entity
	Start []float64 // 10, 20, 30
	End   []float64 // 11, 21, 31
}

func (l *Line) IsEntity() bool {
	return true
}

func NewLine() *Line {
	l := &Line{
		entity: NewEntity(LINE),
		Start:  []float64{0.0, 0.0, 0.0},
		End:    []float64{0.0, 0.0, 0.0},
	}
	return l
}

func ParseLine(data [][2]string) (Entity, error) {
	l := NewLine()
	for _, d := range data {
		switch d[0] {
		case "0":
		case "5":
		case "8":
			// TODO: set layer
		case "10":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.Start[0] = val
		case "20":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.Start[1] = val
		case "30":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.Start[2] = val
		case "11":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.End[0] = val
		case "21":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.End[1] = val
		case "31":
			val, err := strconv.ParseFloat(d[1], 64)
			if err != nil {
				return l, fmt.Errorf("code %s: %s", d[0], err.Error())
			}
			l.End[2] = val
		}
	}
	return l, nil
}

func (l *Line) Format(f *format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLine")
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, l.Start[i])
	}
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10+1, l.End[i])
	}
}

func (l *Line) String() string {
	f := format.New()
	return l.FormatString(f)
}

func (l *Line) FormatString(f *format.Formatter) string {
	l.Format(f)
	return f.Output()
}
