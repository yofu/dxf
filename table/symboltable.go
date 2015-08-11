package table

type SymbolTable interface {
	IsSymbolTable() bool
	String() string
	Handle() int
	SetHandle(*int)
}
