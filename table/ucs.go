package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type Ucs struct {
	handle int
	owner  handle.Handler
	name   string // 2
}

func NewUCS(name string) *Ucs {
	u := new(Ucs)
	u.name = name
	return u
}

func (u *Ucs) IsSymbolTable() bool {
	return true
}

func (u *Ucs) Format(f *format.Formatter) {
	f.WriteString(0, "UCS")
	f.WriteHex(5, u.handle)
	if u.owner != nil {
		f.WriteHex(330, u.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbostableRecord")
	f.WriteString(100, "AcDbUCSTableRecord")
	f.WriteString(2, u.name)
}

func (u *Ucs) String() string {
	f := format.New()
	return u.FormatString(f)
}

func (u *Ucs) FormatString(f *format.Formatter) string {
	u.Format(f)
	return f.Output()
}

func (u *Ucs) Handle() int {
	return u.handle
}
func (u *Ucs) SetHandle(h *int) {
	u.handle = *h
	(*h)++
}

func (u *Ucs) SetOwner(h handle.Handler) {
	u.owner = h
}

func (u *Ucs) Name() string {
	return u.name
}

