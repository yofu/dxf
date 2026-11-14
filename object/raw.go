package object

import (
	"strconv"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Raw is an object of the OBJECTS section that this package does not
// interpret (e.g. XRECORD, MLINESTYLE, IMAGEDEF). Its group codes are kept
// as read and written back with the same handle, so that references by
// handle to and from it stay valid.
type Raw struct {
	Type   string      // 0
	Data   [][2]string // group codes except 0 and 5
	handle int
	src    int
}

// IsObject is for Object interface.
func (r *Raw) IsObject() bool {
	return true
}

// NewRaw creates a Raw object with its handle h in the file it is read from.
func NewRaw(typ, h string, data [][2]string) *Raw {
	return &Raw{Type: typ, Data: data, src: handle.Parse(h)}
}

// Format writes data to formatter.
func (r *Raw) Format(f format.Formatter) {
	f.WriteString(0, r.Type)
	f.WriteHex(5, r.handle)
	for _, p := range r.Data {
		if c, err := strconv.Atoi(p[0]); err == nil {
			f.WriteString(c, p[1])
		}
	}
}

// Handle returns the handle.
func (r *Raw) Handle() int {
	return r.handle
}

// SetHandle sets the handle (the handle read is kept).
func (r *Raw) SetHandle(v *int) {
	r.handle = handle.Next(r.src, v)
}

// SourceHandle returns the handle in the file it was read from.
func (r *Raw) SourceHandle() string {
	return handle.Format(r.src)
}
