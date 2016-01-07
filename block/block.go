package block

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/table"
)

// Block represents each BLOCK.
type Block struct {
	Name        string
	Description string
	handle      int
	endhandle   int
	layer       *table.Layer
	Flag        int
	Coord       []float64
}

// NewBlock create a new Block.
func NewBlock(name, desc string) *Block {
	b := &Block{
		Name:        name,
		Description: desc,
		handle:      0,
		layer:       table.LY_0,
		Flag:        0,
		Coord:       []float64{0.0, 0.0, 0.0},
	}
	return b
}

// Format writes data to formatter.
func (b *Block) Format(f format.Formatter) {
	f.WriteString(0, "BLOCK")
	f.WriteHex(5, b.handle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, b.layer.Name())
	f.WriteString(100, "AcDbBlockBegin")
	f.WriteString(2, b.Name)
	f.WriteInt(70, b.Flag)
	for i := 0; i < 3; i++ {
		f.WriteFloat((i+1)*10, b.Coord[i])
	}
	f.WriteString(3, b.Name)
	f.WriteString(1, b.Description)
	f.WriteString(0, "ENDBLK")
	f.WriteHex(5, b.endhandle)
	f.WriteString(100, "AcDbEntity")
	f.WriteString(8, b.layer.Name())
	f.WriteString(100, "AcDbBlockEnd")
}

// String outputs data using default formatter.
func (b *Block) String() string {
	f := format.NewASCII()
	return b.FormatString(f)
}

// FormatString outputs data using given formatter.
func (b *Block) FormatString(f format.Formatter) string {
	b.Format(f)
	return f.Output()
}

// Handle returns a handle value of BLOCK.
func (b *Block) Handle() int {
	return b.handle
}

// SetHandle sets handles to BLOCK and ENDBLK.
func (b *Block) SetHandle(v *int) {
	b.handle = *v
	(*v)++
	b.endhandle = *v
	(*v)++
}

// Layer returns BLOCK's Layer.
func (b *Block) Layer() *table.Layer {
	return b.layer
}

// SetLayer sets Layer to BLOCK.
func (b *Block) SetLayer(layer *table.Layer) {
	b.layer = layer
}
