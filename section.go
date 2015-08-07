package dxf

import (
	"bytes"
)

type Section interface {
	WriteTo(*bytes.Buffer) error
	SetHandle(*int)
}
