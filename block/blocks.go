package block

import (
	"bytes"
)

type Blocks struct {
}

func New() *Blocks {
	b := new(Blocks)
	return b
}

func (b *Blocks) WriteTo(otp *bytes.Buffer) error {
	otp.WriteString("0\nSECTION\n2\nBLOCKS\n")
	otp.WriteString("0\nENDSEC\n")
	return nil
}

func (bs *Blocks) SetHandle(v *int) {
	return
}
