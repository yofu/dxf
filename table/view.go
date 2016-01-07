package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// View represents VIEW SymbolTable.
type View struct {
	handle int
	owner  handle.Handler
	name   string // 2
}

// NewView creates a new View.
func NewView(name string) *View {
	v := new(View)
	v.name = name
	return v
}

// IsSymbolTable is for SymbolTable interface.
func (v *View) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (v *View) Format(f format.Formatter) {
	f.WriteString(0, "VIEW")
	f.WriteHex(5, v.handle)
	if v.owner != nil {
		f.WriteHex(330, v.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbViewTableRecord")
	f.WriteString(2, v.name)
}

// String outputs data using default formatter.
func (v *View) String() string {
	f := format.NewASCII()
	return v.FormatString(f)
}

// FormatString outputs data using given formatter.
func (v *View) FormatString(f format.Formatter) string {
	v.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (v *View) Handle() int {
	return v.handle
}

// SetHandle sets a handle.
func (v *View) SetHandle(h *int) {
	v.handle = *h
	(*h)++
}

// SetOwner sets an owner.
func (v *View) SetOwner(h handle.Handler) {
	v.owner = h
}

// Name returns a name of VIEW (code 2).
func (v *View) Name() string {
	return v.name
}
