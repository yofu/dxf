package entity

import (
	"strconv"
	"strings"
)

// XValue is a typed value of extended data (XDATA): Code is the group
// code (1000 string, 1002 "{"/"}" list, 1003 layer, 1004 binary, 1005
// handle, 1010 point, 1011-1013 world point/displacement/direction, 1040
// real, 1041 distance, 1042 scale, 1070 int16, 1071 int32) and Value a
// string, float64, int or []float64 (points).
type XValue struct {
	Code  int
	Value interface{}
}

// XString, XReal, XInt, XInt32, XPoint and XList build typed values.
func XString(s string) XValue        { return XValue{1000, s} }
func XReal(v float64) XValue         { return XValue{1040, v} }
func XInt(v int) XValue              { return XValue{1070, v} }
func XInt32(v int) XValue            { return XValue{1071, v} }
func XPoint(p []float64) XValue      { return XValue{1010, p} }
func XHandle(h string) XValue        { return XValue{1005, h} }
func XListBegin() XValue             { return XValue{1002, "{"} }
func XListEnd() XValue               { return XValue{1002, "}"} }
func XLayer(name string) XValue      { return XValue{1003, name} }
func XDistance(v float64) XValue     { return XValue{1041, v} }
func XScale(v float64) XValue        { return XValue{1042, v} }
func XWorldPoint(p []float64) XValue { return XValue{1011, p} }

func (v XValue) pairs() [][2]string {
	c := strconv.Itoa(v.Code)
	switch x := v.Value.(type) {
	case []float64:
		return [][2]string{{c, ftoa(coord(x, 0))}, {strconv.Itoa(v.Code + 10), ftoa(coord(x, 1))}, {strconv.Itoa(v.Code + 20), ftoa(coord(x, 2))}}
	case float64:
		return [][2]string{{c, ftoa(x)}}
	case int:
		return [][2]string{{c, strconv.Itoa(x)}}
	case string:
		return [][2]string{{c, x}}
	}
	return nil
}

// XDataValues returns the typed values of the extended data of
// application app (nil if there is none).
func (e *entity) XDataValues(app string) []XValue {
	var out []XValue
	in := false
	for i := 0; i < len(e.xdata); i++ {
		p := e.xdata[i]
		code, _ := strconv.Atoi(strings.TrimSpace(p[0]))
		if code == 1001 {
			in = strings.EqualFold(strings.TrimSpace(p[1]), app)
			continue
		}
		if !in {
			continue
		}
		v := strings.TrimSpace(p[1])
		switch {
		case code >= 1010 && code <= 1013:
			pt := []float64{0, 0, 0}
			pt[0], _ = strconv.ParseFloat(v, 64)
			for k := 1; k <= 2 && i+1 < len(e.xdata); k++ {
				c, _ := strconv.Atoi(strings.TrimSpace(e.xdata[i+1][0]))
				if c != code+10*k {
					break
				}
				i++
				pt[k], _ = strconv.ParseFloat(strings.TrimSpace(e.xdata[i][1]), 64)
			}
			out = append(out, XValue{code, pt})
		case code >= 1040 && code <= 1042:
			f, _ := strconv.ParseFloat(v, 64)
			out = append(out, XValue{code, f})
		case code == 1070 || code == 1071:
			n, _ := strconv.Atoi(v)
			out = append(out, XValue{code, n})
		default:
			out = append(out, XValue{code, p[1]})
		}
	}
	return out
}

// SetXDataValues sets the extended data of application app to values
// (replacing its data; other applications are kept). The application is
// added to the APPID table when the drawing is written.
func (e *entity) SetXDataValues(app string, values ...XValue) {
	var out [][2]string
	skip := false
	for _, p := range e.xdata {
		if strings.TrimSpace(p[0]) == "1001" {
			skip = strings.EqualFold(strings.TrimSpace(p[1]), app)
		}
		if !skip {
			out = append(out, p)
		}
	}
	out = append(out, [2]string{"1001", app})
	for _, v := range values {
		out = append(out, v.pairs()...)
	}
	e.xdata = out
}

// XDataApps returns the applications of the extended data.
func (e *entity) XDataApps() []string {
	var apps []string
	for _, p := range e.xdata {
		if strings.TrimSpace(p[0]) == "1001" {
			apps = append(apps, strings.TrimSpace(p[1]))
		}
	}
	return apps
}
