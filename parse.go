package dxf

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const entityTypeCode = "0"

// setFloat sets a floating point number to a variable using given function.
func setFloat(data [2]string, f func(float64)) error {
	val, err := strconv.ParseFloat(strings.TrimSpace(data[1]), 64)
	if err != nil {
		return fmt.Errorf("code %s: %w", data[0], err)
	}
	f(val)
	return nil
}

// setInt sets an integer value to a variable using given function.
func setInt(data [2]string, f func(int)) error {
	val, err := strconv.ParseInt(strings.TrimSpace(data[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("code %s: %w", data[0], err)
	}
	f(int(val))
	return nil
}

// HEADER

// CLASSES

// setSourceHandle keeps the handle (5, or 105 for DIMSTYLE) of an object read from a file.
func setSourceHandle(o interface{}, data [][2]string) {
	sh, ok := o.(interface{ SetSourceHandle(string) })
	if !ok {
		return
	}
	for _, code := range []string{"5", "105"} {
		if h, ok := value(data, code); ok {
			sh.SetSourceHandle(h)
			return
		}
	}
}

// setExtrusion sets a coordinate of an extrusion direction (210, 220, 230).
func setExtrusion(dt [2]string, v []float64) error {
	i := int(dt[0][1] - '1')
	if i < 0 || i >= len(v) {
		return nil
	}
	return setFloat(dt, func(val float64) { v[i] = val })
}

// value returns the value of the first group code in data.
func value(data [][2]string, code string) (string, bool) {
	for _, dt := range data {
		if dt[0] == code {
			return dt[1], true
		}
	}
	return "", false
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// codes returns a set of group codes.
func codes(cs ...string) map[string]bool {
	m := make(map[string]bool, len(cs))
	for _, c := range cs {
		m[c] = true
	}
	return m
}
