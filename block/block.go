package block

import (
	"bytes"
	"fmt"
	"github.com/yofu/dxf/table"
)

type Block struct {
	Name        string
	Description string
	handle      int
	endhandle   int
	layer       *table.Layer
	Flag        int
	Coord       []float64
}

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

func (b *Block) String() string {
	var otp bytes.Buffer
	otp.WriteString("0\nBLOCK\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", b.handle))
	otp.WriteString("100\nAcDbEntity\n")
	otp.WriteString(fmt.Sprintf("8\n%s\n", b.layer.Name))
	otp.WriteString("100\nAcDbBlockBegin\n")
	otp.WriteString(fmt.Sprintf("2\n%s\n", b.Name))
	otp.WriteString(fmt.Sprintf("70\n%d\n", b.Flag))
	for i := 0; i < 3; i++ {
		otp.WriteString(fmt.Sprintf("%d\n%f\n", (i+1)*10, b.Coord[i]))
	}
	otp.WriteString(fmt.Sprintf("3\n%s\n", b.Name))
	otp.WriteString(fmt.Sprintf("1\n%s\n", b.Description))
	otp.WriteString("0\nENDBLK\n")
	otp.WriteString(fmt.Sprintf("5\n%x\n", b.endhandle))
	otp.WriteString("100\nAcDbEntity\n")
	otp.WriteString(fmt.Sprintf("8\n%s\n", b.layer.Name))
	otp.WriteString("100\nAcDbBlockEnd\n")
	return otp.String()
}

func (b *Block) Handle() int {
	return b.handle
}
func (b *Block) SetHandle(v *int) {
	b.handle = *v
	(*v)++
	b.endhandle = *v
	(*v)++
}
