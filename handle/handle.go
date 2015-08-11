package handle

type Handler interface {
	Handle() int
	SetHandle(*int)
}
