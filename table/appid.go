package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// AppID represents APPID SymbolTable.
type AppID struct {
	handle int
	owner  handle.Handler
	name   string
}

// NewAppID create a new AppID.
func NewAppID(name string) *AppID {
	a := new(AppID)
	a.name = name
	return a
}

// IsSymbolTable is for SymbolTable interface.
func (a *AppID) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (a *AppID) Format(f format.Formatter) {
	f.WriteString(0, "APPID")
	f.WriteHex(5, a.handle)
	if a.owner != nil {
		f.WriteHex(330, a.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbRegAppTableRecord")
	f.WriteString(2, a.name)
	f.WriteInt(70, 0)
}

// String outputs data using default formatter.
func (a *AppID) String() string {
	f := format.NewASCII()
	return a.FormatString(f)
}

// FormatString outputs data using given formatter.
func (a *AppID) FormatString(f format.Formatter) string {
	a.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (a *AppID) Handle() int {
	return a.handle
}

// SetHandle sets a handle.
func (a *AppID) SetHandle(v *int) {
	a.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (a *AppID) SetOwner(h handle.Handler) {
	a.owner = h
}

// Name returns a name of APPID (code 2).
func (a *AppID) Name() string {
	return a.name
}
