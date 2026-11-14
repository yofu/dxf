package dxf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/header"
	"github.com/yofu/dxf/insunit"
)

// headerVars returns setters for the header variables supported by this
// package, keyed by variable name and group code.
func headerVars(h *header.Header) map[string]map[string]func(string) error {
	f := func(p *float64) func(string) error {
		return func(v string) error {
			x, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err == nil {
				*p = x
			}
			return err
		}
	}
	i := func(p *int) func(string) error {
		return func(v string) error {
			x, err := strconv.Atoi(strings.TrimSpace(v))
			if err == nil {
				*p = x
			}
			return err
		}
	}
	s := func(p *string) func(string) error {
		return func(v string) error {
			*p = strings.TrimSpace(v)
			return nil
		}
	}
	xyz := func(p []float64) map[string]func(string) error {
		m := map[string]func(string) error{"10": f(&p[0]), "20": f(&p[1])}
		if len(p) > 2 {
			m["30"] = f(&p[2])
		}
		return m
	}
	return map[string]map[string]func(string) error{
		"$ACADVER":     {"1": func(v string) error { h.Version = v; return nil }},
		"$DWGCODEPAGE": {"3": s(&h.CodePage)},
		"$INSBASE":     xyz(h.InsBase),
		"$EXTMIN":      xyz(h.ExtMin),
		"$EXTMAX":      xyz(h.ExtMax),
		"$LIMMIN":      xyz(h.LimMin),
		"$LIMMAX":      xyz(h.LimMax),
		"$LTSCALE":     {"40": f(&h.LtScale)},
		"$INSUNITS": {"70": func(v string) error {
			x, err := strconv.Atoi(strings.TrimSpace(v))
			h.InsUnit = insunit.Unit(x)
			return err
		}},
		"$LUNITS": {"70": func(v string) error {
			// Need to adjust the value back to the constants
			x, err := strconv.Atoi(strings.TrimSpace(v))
			h.InsLUnit = insunit.Type(x - 2)
			return err
		}},
		"$LUPREC":      {"70": i(&h.LUPrec)},
		"$TEXTSIZE":    {"40": f(&h.TextSize)},
		"$CLAYER":      {"8": s(&h.CLayer)},
		"$CELTYPE":     {"6": s(&h.CELType)},
		"$CECOLOR":     {"62": i(&h.CEColor)},
		"$CELTSCALE":   {"40": f(&h.CELTScale)},
		"$CELWEIGHT":   {"370": i(&h.CELWeight)},
		"$DIMSTYLE":    {"2": s(&h.DimStyle)},
		"$TEXTSTYLE":   {"7": s(&h.TextStyle)},
		"$MEASUREMENT": {"70": i(&h.Measurement)},
		"$AUNITS":      {"70": i(&h.AUnits)},
		"$AUPREC":      {"70": i(&h.AUPrec)},
		"$ANGBASE":     {"50": f(&h.AngBase)},
		"$ANGDIR":      {"70": i(&h.AngDir)},
		"$PDMODE":      {"70": i(&h.PdMode)},
		"$PDSIZE":      {"40": f(&h.PdSize)},
	}
}

// ParseHeader parses HEADER section.
func ParseHeader(d *drawing.Drawing, line int, data [][2]string) error {
	h := d.Header()
	vars := headerVars(h)
	h.Extra = nil
	var name string
	for _, dt := range data {
		if dt[0] == "9" {
			name = strings.TrimSpace(dt[1])
			if _, known := vars[name]; !known && name != "$HANDSEED" {
				h.Extra = append(h.Extra, [2]string{"9", name})
			}
			continue
		}
		if set, ok := vars[name][dt[0]]; ok {
			if err := set(dt[1]); err != nil {
				return fmt.Errorf("%s code %s: %w", name, dt[0], err)
			}
			continue
		}
		if _, known := vars[name]; known || name == "$HANDSEED" || name == "" || isPointer(dt[0]) {
			continue
		}
		// variables not interpreted by this package are kept as read
		h.Extra = append(h.Extra, [2]string{strings.TrimSpace(dt[0]), dt[1]})
	}
	// drop variables left without values (e.g. only handle references)
	var extra [][2]string
	for i, p := range h.Extra {
		if p[0] == "9" && (i+1 == len(h.Extra) || h.Extra[i+1][0] == "9") {
			continue
		}
		extra = append(extra, p)
	}
	h.Extra = extra
	return nil
}
