package table

import (
	"bytes"
	"fmt"
)

type BlockRecord struct {
	handle   int
	Name     string
}

func NewBlockRecord(name string) *BlockRecord {
	b := new(BlockRecord)
	b.Name = name
	return b
}

func (b *BlockRecord) IsSymbolTable() bool {
	return true
}

func (b *BlockRecord) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nBLOCK_RECORD\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", b.handle))
	otp.WriteString("100\nAcDbSymbolTableRecord\n100\nAcDbBlockTableRecord\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", b.Name))
	otp.WriteString("70\n0\n")
	otp.WriteString("280\n1\n")
	otp.WriteString("281\n0\n")
	return otp.String()
}

func (b *BlockRecord) Handle() int {
	return b.handle
}
func (b *BlockRecord) SetHandle(v *int) {
	b.handle = *v
	(*v)++
}
