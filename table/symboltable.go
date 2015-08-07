package table

type SymbolTable interface {
	IsSymbolTable() bool
	String() string
	SetHandle(*int)
}
