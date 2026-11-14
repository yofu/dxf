package object

import (
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Raw is an object of the OBJECTS section kept as its group codes: the
// objects this package does not interpret, and the base of the typed
// objects (XRecord, DictionaryVar, ...), whose accessors read and change
// the group codes. They are written back with the same handle, so that
// references by handle to and from them stay valid.
type Raw struct {
	Type   string      // 0
	Data   [][2]string // group codes except 0 and 5
	handle int
	src    int
	refs   map[int]handle.Handler // Data index -> object written with its handle
}

// IsObject is for Object interface.
func (r *Raw) IsObject() bool {
	return true
}

// NewRaw creates a Raw object with its handle h in the file it is read from.
func NewRaw(typ, h string, data [][2]string) *Raw {
	return &Raw{Type: typ, Data: data, src: handle.Parse(h)}
}

// newObject creates an object with owner (written as reactor and 330)
// and the group codes data after the owner.
func newObject(typ string, owner handle.Handler, data ...[2]string) *Raw {
	r := &Raw{Type: typ, Data: [][2]string{{"102", "{ACAD_REACTORS"}, {"330", "0"}, {"102", "}"}, {"330", "0"}}}
	r.Data = append(r.Data, data...)
	if owner != nil {
		r.SetOwner(owner)
	}
	return r
}

// Format writes data to formatter.
func (r *Raw) Format(f format.Formatter) {
	f.WriteString(0, r.Type)
	f.WriteHex(5, r.handle)
	for i, p := range r.Data {
		c, err := strconv.Atoi(strings.TrimSpace(p[0]))
		if err != nil {
			continue
		}
		if h, ok := r.refs[i]; ok && h != nil {
			f.WriteHex(c, h.Handle())
			continue
		}
		f.WriteString(c, p[1])
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

// RawObject returns the object itself (also for the typed objects based on
// Raw).
func (r *Raw) RawObject() *Raw { return r }

// SetRef makes the value at Data index i the handle of h when written.
func (r *Raw) SetRef(i int, h handle.Handler) {
	if r.refs == nil {
		r.refs = map[int]handle.Handler{}
	}
	r.refs[i] = h
}

// Ref returns the object referred to at Data index i (nil if the value is
// a handle as read).
func (r *Raw) Ref(i int) handle.Handler { return r.refs[i] }

// AppendRef appends code with the handle of h.
func (r *Raw) AppendRef(code string, h handle.Handler) {
	r.Data = append(r.Data, [2]string{code, "0"})
	r.SetRef(len(r.Data)-1, h)
}

// SetOwner sets the owner: the first 330 in the reactors group and the
// first 330 after it (the owner of the object).
func (r *Raw) SetOwner(h handle.Handler) {
	inReactors, doneReactor := false, false
	for i, p := range r.Data {
		switch {
		case p[0] == "102" && strings.EqualFold(strings.TrimSpace(p[1]), "{ACAD_REACTORS"):
			inReactors = true
		case p[0] == "102":
			inReactors = false
		case p[0] == "330" && inReactors && !doneReactor:
			r.SetRef(i, h)
			doneReactor = true
		case p[0] == "330" && !inReactors:
			r.SetRef(i, h)
			return
		case p[0] == "100":
			return
		}
	}
}

// Owner returns the owner set by SetOwner or the handle as read.
func (r *Raw) Owner() (handle.Handler, string) {
	in := false
	for i, p := range r.Data {
		switch {
		case p[0] == "102":
			in = strings.HasPrefix(p[1], "{")
		case p[0] == "330" && !in:
			return r.refs[i], strings.TrimSpace(p[1])
		case p[0] == "100":
			return nil, ""
		}
	}
	return nil, ""
}

// Value returns the first value of code after the subclass markers
// (100) and the owner.
func (r *Raw) Value(code string) (string, bool) {
	for _, p := range r.body() {
		if p[0] == code {
			return p[1], true
		}
	}
	return "", false
}

// Float returns the first value of code as float (def if missing).
func (r *Raw) Float(code string, def float64) float64 {
	v, ok := r.Value(code)
	if !ok {
		return def
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return def
	}
	return f
}

// Int returns the first value of code as int (def if missing).
func (r *Raw) Int(code string, def int) int {
	v, ok := r.Value(code)
	if !ok {
		return def
	}
	i, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return i
}

// Set sets the first value of code after the first subclass marker, or
// appends it.
func (r *Raw) Set(code, value string) {
	start := r.bodyStart()
	for i := start; i < len(r.Data); i++ {
		if r.Data[i][0] == code {
			r.Data[i][1] = value
			delete(r.refs, i)
			return
		}
	}
	r.Data = append(r.Data, [2]string{code, value})
}

// SetFloat sets the first value of code (see Set).
func (r *Raw) SetFloat(code string, v float64) {
	r.Set(code, strconv.FormatFloat(v, 'f', -1, 64))
}

// SetInt sets the first value of code (see Set).
func (r *Raw) SetInt(code string, v int) { r.Set(code, strconv.Itoa(v)) }

// bodyStart returns the index of the first group code after the first
// subclass marker (len(Data) if there is none).
func (r *Raw) bodyStart() int {
	for i, p := range r.Data {
		if p[0] == "100" {
			return i + 1
		}
	}
	return len(r.Data)
}

func (r *Raw) body() [][2]string { return r.Data[r.bodyStart():] }
