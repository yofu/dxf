package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// BlockRecord represents BLOCK_RECORD SymbolTable.
type BlockRecord struct {
	handle int
	owner  handle.Handler
	name   string
	layout handle.Handler // 340: LAYOUT object (model space and paper space)

	// SourceHandle is the handle (code 5) in the file the record was read from.
	SourceHandle string

	InsUnits   int  // 70: insertion units
	Explodable bool // 280
	Scalable   bool // 281: uniform scaling only if false
	// Extra holds group codes read from a file that are not interpreted by
	// this package (and XDATA); they are written after the other values.
	Extra [][2]string
}

// NewBlockRecord creates a new BlockRecord.
func NewBlockRecord(name string) *BlockRecord {
	b := new(BlockRecord)
	b.name = name
	b.Explodable = true
	return b
}

// IsSymbolTable is for SymbolTable interface.
func (b *BlockRecord) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (b *BlockRecord) Format(f format.Formatter) {
	f.WriteString(0, "BLOCK_RECORD")
	f.WriteHex(5, b.handle)
	if b.owner != nil {
		f.WriteHex(330, b.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbBlockTableRecord")
	f.WriteString(2, b.name)
	if b.layout != nil {
		f.WriteHex(340, b.layout.Handle())
	}
	f.WriteInt(70, b.InsUnits)
	f.WriteInt(280, boolInt(b.Explodable))
	f.WriteInt(281, boolInt(b.Scalable))
	writeRaw(f, b.Extra)
}

// String outputs data using default formatter.
func (b *BlockRecord) String() string {
	f := format.NewASCII()
	return b.FormatString(f)
}

// FormatString outputs data using given formatter.
func (b *BlockRecord) FormatString(f format.Formatter) string {
	b.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (b *BlockRecord) Handle() int {
	return b.handle
}

// SetHandle sets a handle.
func (b *BlockRecord) SetHandle(v *int) {
	b.handle = *v
	(*v)++
}

// SetOwner sets an owner.
func (b *BlockRecord) SetOwner(h handle.Handler) {
	b.owner = h
}

// Name returns a name of BLOCK_RECORD (code 2).
func (b *BlockRecord) Name() string {
	return b.name
}

// Layout returns the associated LAYOUT object (nil for ordinary blocks).
func (b *BlockRecord) Layout() handle.Handler {
	return b.layout
}

// SetLayout sets the associated LAYOUT object (code 340).
func (b *BlockRecord) SetLayout(l handle.Handler) {
	b.layout = l
}

// SetName sets the name of the block record.
func (b *BlockRecord) SetName(name string) {
	b.name = name
}
