package block

import (
	"github.com/yofu/dxf/format"
)

type Blocks []*Block

func New() Blocks {
	b := make([]*Block, 3)
	b[0] = NewBlock("*Model_Space", "")
	b[1] = NewBlock("*Paper_Space", "")
	b[2] = NewBlock("*Paper_Space0", "")
	return b
}

func (bs Blocks) WriteTo(f *format.Formatter) {
	f.WriteString(0, "SECTION")
	f.WriteString(2, "BLOCKS")
	for _, b := range bs {
		b.Format(f)
	}
	f.WriteString(0, "ENDSEC")
}

func (bs Blocks) Add(b *Block) Blocks {
	bs = append(bs, b)
	return bs
}

func (bs Blocks) SetHandle(v *int) {
	for _, b := range bs {
		b.SetHandle(v)
	}
}

func (bs Blocks) Read(line int, data [][2]string) error {
	return nil
}
