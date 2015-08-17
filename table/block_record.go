package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type BlockRecord struct {
	handle   int
	owner    handle.Handler
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

func (b *BlockRecord) Format(f *format.Formatter) {
	f.WriteString(0, "BLOCK_RECORD")
	f.WriteHex(5, b.handle)
	if b.owner != nil {
		f.WriteHex(330, b.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbBlockTableRecord")
	f.WriteString(2, b.Name)
	f.WriteInt(70, 0)
	f.WriteInt(280, 1)
	f.WriteInt(281, 0)
}

func (b *BlockRecord) String() string {
	f := format.New()
	return b.FormatString(f)
}

func (b *BlockRecord) FormatString(f *format.Formatter) string {
	b.Format(f)
	return f.Output()
}

func (b *BlockRecord) Handle() int {
	return b.handle
}
func (b *BlockRecord) SetHandle(v *int) {
	b.handle = *v
	(*v)++
}

func (b *BlockRecord) SetOwner(h handle.Handler) {
	b.owner = h
}
