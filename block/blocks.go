// BLOCK section
package block

import (
	"github.com/scantrust/dxf-golang/format"
)

// Blocks represents BLOCKS section.
type Blocks []*Block

// New creates a new Blocks.
func New() Blocks {
	b := make([]*Block, 3)
	b[0] = NewBlock("*Model_Space", "")
	b[1] = NewBlock("*Paper_Space", "")
	b[2] = NewBlock("*Paper_Space0", "")
	return b
}

// Format writes BLOCKS data to formatter.
func (bs Blocks) Format(f format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "BLOCKS")
	for _, b := range bs {
		b.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

// Add adds a new block to BLOCKS section.
func (bs Blocks) Add(b *Block) Blocks {
	bs = append(bs, b)
	return bs
}

// SetHandle sets handles to each block.
func (bs Blocks) SetHandle(v *int) {
	for _, b := range bs {
		b.SetHandle(v)
	}
}
