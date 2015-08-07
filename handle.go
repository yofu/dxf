package dxf

type Handler interface {
	Handle() int
	SetHandle(*int)
}
