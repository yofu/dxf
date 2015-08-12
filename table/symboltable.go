package table

import (
	"github.com/yofu/dxf/handle"
)

type SymbolTable interface {
	IsSymbolTable() bool
	String() string
	Handle() int
	SetHandle(*int)
	SetOwner(handle.Handler)
}
