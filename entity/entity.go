package entity

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/handle"
	"github.com/yofu/dxf/table"
)

type Entity interface {
	IsEntity() bool
	String() string
	Handle() int
	SetHandle(*int)
	SetBlockRecord(handle.Handler)
	Layer() *table.Layer
	SetLayer(*table.Layer)
}

type entity struct {
	Type        EntityType     // 0
	handle      int            // 5
	blockRecord handle.Handler // 330
	layer       *table.Layer   // 8
}

func NewEntity(t EntityType) *entity {
	e := new(entity)
	e.Type = t
	e.layer = table.LY_0
	return e
}

func (e *entity) String() string {
	var otp bytes.Buffer
	otp.WriteString(fmt.Sprintf("0\n%s\n", EntityTypeString(e.Type)))
	otp.WriteString(fmt.Sprintf("5\n%x\n", e.handle))
	if e.blockRecord != nil {
		otp.WriteString(fmt.Sprintf("102\n{ACAD_REACTORS\n330\n%x\n102\n}\n", e.blockRecord.Handle()))
	}
	otp.WriteString("100\nAcDbEntity\n")
	otp.WriteString(fmt.Sprintf("8\n%s\n", e.layer.Name))
	return otp.String()
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

func (e *entity) Layer() *table.Layer {
	return e.layer
}

func (e *entity) SetLayer(l *table.Layer) {
	e.layer = l
}
