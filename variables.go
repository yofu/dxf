package dxf

import (
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/format"
)

// HeaderVariable returns the group codes and values of the header
// variable name (e.g. "$PDMODE"), as written: the variables this package
// interprets and those kept as read (Header().Extra).
func HeaderVariable(d *drawing.Drawing, name string) ([][2]string, bool) {
	name = strings.ToUpper(strings.TrimSpace(name))
	f := format.NewASCII()
	f.SetPrecision(16)
	d.Header().Format(f)
	lines := strings.Split(strings.ReplaceAll(f.Output(), "\r\n", "\n"), "\n")
	var out [][2]string
	found := false
	for i := 0; i+1 < len(lines); i += 2 {
		code, value := strings.TrimSpace(lines[i]), lines[i+1]
		if code == "9" {
			if found {
				break
			}
			found = strings.EqualFold(strings.TrimSpace(value), name)
			continue
		}
		if code == "0" && found {
			break
		}
		if found {
			out = append(out, [2]string{code, value})
		}
	}
	return out, found
}

// SetHeaderVariable sets the header variable name to the group codes and
// values (e.g. {"70", "3"}): the variables this package interprets are
// set in Header() (as when reading), the others are kept in Header().Extra
// (replaced or added).
func SetHeaderVariable(d *drawing.Drawing, name string, values ...[2]string) error {
	name = strings.ToUpper(strings.TrimSpace(name))
	h := d.Header()
	if setters, ok := headerVars(h)[name]; ok {
		for _, v := range values {
			if set, ok := setters[v[0]]; ok {
				if err := set(strings.TrimSpace(v[1])); err != nil {
					return err
				}
			}
		}
		return nil
	}
	var extra [][2]string
	replaced := false
	skip := false
	for _, p := range h.Extra {
		if p[0] == "9" {
			skip = strings.EqualFold(strings.TrimSpace(p[1]), name)
			if skip && !replaced {
				extra = append(extra, [2]string{"9", name})
				extra = append(extra, values...)
				replaced = true
			}
			if skip {
				continue
			}
		}
		if !skip {
			extra = append(extra, p)
		}
	}
	if !replaced {
		extra = append(extra, [2]string{"9", name})
		extra = append(extra, values...)
	}
	h.Extra = extra
	return nil
}
