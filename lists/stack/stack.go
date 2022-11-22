package stack

type Node struct {
	prev  *Node
	Value any
}

type Stack struct {
	node *Node
}

func New() *Stack {
	node := &Node{}
	node.prev = node
	return &Stack{
		node: node,
	}
}

func (s *Stack) IsEmpty() bool {
	return s.node == s.node.prev
}

func (s *Stack) Push(val any) {
	s.node = &Node{
		prev:  s.node,
		Value: val,
	}
}

func (s *Stack) Pop() (val any, ok bool) {
	node := s.node
	s.node = node.prev
	return node.Value, node != s.node
}
