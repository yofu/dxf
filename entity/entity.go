package entity

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/table"
)

type Entity interface {
	IsEntity() bool
	String() string
	SetHandle(*int)
	Layer() *table.Layer
	SetLayer(*table.Layer)
}

type entity struct {
	Type   EntityType // 0
	handle int        // 5
	layer  *table.Layer     // 8
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

func (e *entity) Layer() *table.Layer {
	return e.layer
}

func (e *entity) SetLayer(l *table.Layer) {
	e.layer = l
}
