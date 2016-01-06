package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type AppID struct {
	handle   int
	owner    handle.Handler
	name     string
}

func NewAppID(name string) *AppID {
	a := new(AppID)
	a.name = name
	return a
}

func (a *AppID) IsSymbolTable() bool {
	return true
}

func (a *AppID) Format(f *format.Formatter) {
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

func (a *AppID) String() string {
	f := format.New()
	return a.FormatString(f)
}

func (a *AppID) FormatString(f *format.Formatter) string {
	a.Format(f)
	return f.Output()
}

func (a *AppID) Handle() int {
	return a.handle
}
func (a *AppID) SetHandle(v *int) {
	a.handle = *v
	(*v)++
}

func (a *AppID) SetOwner(h handle.Handler) {
	a.owner = h
}

func (a *AppID) Name() string {
	return a.name
}
