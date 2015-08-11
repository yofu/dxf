package block

import (
	"bytes"
)

type Blocks []*Block

func New() Blocks {
	b := make([]*Block, 3)
	b[0] = NewBlock("*Model_Space", "")
	b[1] = NewBlock("*Paper_Space", "")
	b[2] = NewBlock("*Paper_Space0", "")
	return b
}

func (bs Blocks) WriteTo(otp *bytes.Buffer) error {
	otp.WriteString("0\nSECTION\n2\nBLOCKS\n")
	for _, b := range bs {
		otp.WriteString(b.String())
	}
	otp.WriteString("0\nENDSEC\n")
	return nil
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
