package dxf

import (
	"github.com/yofu/dxf/format"
)

type Section interface {
	WriteTo(*format.Formatter)
	SetHandle(*int)
}
