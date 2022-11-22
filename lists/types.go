package lists

type Pushable interface {
	IsEmpty() bool
	Push(val any)
	Pop() (val any, ok bool)
}

type Bidirectional interface {
	Pushable
	PushBack(val any)
}
