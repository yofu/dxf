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
}

type entity struct {
	Type   EntityType // 0
	handle int        // 5
	Layer  *table.Layer     // 8
}

func NewEntity() *entity {
	e := new(entity)
	e.Layer = table.LY_0
	return e
}

func (e *entity) String() string {
	var otp bytes.Buffer
	otp.WriteString(fmt.Sprintf("0\n%s\n", EntityTypeString(e.Type)))
	otp.WriteString(fmt.Sprintf("5\n%x\n", e.handle))
	otp.WriteString("100\nAcDbEntity\n")
	otp.WriteString(fmt.Sprintf("8\n%s\n", e.Layer.Name))
	otp.WriteString("370\n0\n")
	return otp.String()
}

func (e *entity) Handle() int {
	return e.handle
}
func (e *entity) SetHandle(v *int) {
	e.handle = *v
	(*v)++
}
