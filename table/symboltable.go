package table

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

type SymbolTable interface {
	IsSymbolTable() bool
	Format(*format.Formatter)
	Handle() int
	SetHandle(*int)
	SetOwner(handle.Handler)
}
