package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// UCS represents UCS SymbolTable.
type Ucs struct {
	handle int
	owner  handle.Handler
	name   string // 2
}

// NewUCS creates a new Ucs.
func NewUCS(name string) *Ucs {
	u := new(Ucs)
	u.name = name
	return u
}

// IsSymbolTable is for SymbolTable interface.
func (u *Ucs) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (u *Ucs) Format(f format.Formatter) {
	f.WriteString(0, "UCS")
	f.WriteHex(5, u.handle)
	if u.owner != nil {
		f.WriteHex(330, u.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbUCSTableRecord")
	f.WriteString(2, u.name)
}

// String outputs data using default formatter.
func (u *Ucs) String() string {
	f := format.NewASCII()
	return u.FormatString(f)
}

// FormatString outputs data using given formatter.
func (u *Ucs) FormatString(f format.Formatter) string {
	u.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (u *Ucs) Handle() int {
	return u.handle
}

// SetHandle sets a handle.
func (u *Ucs) SetHandle(h *int) {
	u.handle = *h
	(*h)++
}

// SetOwner sets an owner.
func (u *Ucs) SetOwner(h handle.Handler) {
	u.owner = h
}

// Name returns a name of UCS (code 2).
func (u *Ucs) Name() string {
	return u.name
}
