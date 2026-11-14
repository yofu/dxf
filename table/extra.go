package table

import (
	"strconv"

	"github.com/yofu/dxf/format"
)

// writeRaw writes group code/value pairs kept as read
// (Extra of table entries: values not interpreted by this package and XDATA).
func writeRaw(f format.Formatter, pairs [][2]string) {
	for _, p := range pairs {
		if c, err := strconv.Atoi(p[0]); err == nil {
			f.WriteString(c, p[1])
		}
	}
}
