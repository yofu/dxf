package entity

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

type Entity interface {
	IsEntity() bool
	Format(*format.Formatter)
	Handle() int
	SetHandle(*int)
	SetBlockRecord(handle.Handler)
	Layer() *table.Layer
	SetLayer(*table.Layer)
}

type entity struct {
	Type        EntityType     // 0
	handle      int            // 5
	blockRecord handle.Handler // 102 330
	owner       handle.Handler // 330
	layer       *table.Layer   // 8
}

func NewEntity(t EntityType) *entity {
	e := &entity{
		Type:        t,
		handle:      0,
		blockRecord: nil,
		owner:       nil,
		layer:       table.LY_0,
	}
	return e
}

func (e *entity) Format(f *format.Formatter) {
	f.WriteString(0, EntityTypeString(e.Type))
	f.WriteHex(5, e.handle)
	if e.blockRecord != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, e.blockRecord.Handle())
		f.WriteString(102, "}")
	}
	if e.owner != nil {
		f.WriteHex(330, e.owner.Handle())
	}
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, e.layer.Name())
}

func (e *entity) String() string {
	f := format.New()
	return e.FormatString(f)
}

func (e *entity) FormatString(f *format.Formatter) string {
	e.Format(f)
	return f.Output()
}

func (e *entity) Handle() int {
	return e.handle
}
func (e *entity) SetHandle(v *int) {
	e.handle = *v
	(*v)++
}

func (e *entity) SetBlockRecord(h handle.Handler) {
	e.blockRecord = h
}

func (e *entity) SetOwner(h handle.Handler) {
	e.owner = h
}

func (e *entity) Layer() *table.Layer {
	return e.layer
}

func (e *entity) SetLayer(l *table.Layer) {
	e.layer = l
}
