package squeue

import (
	"sync"
)

type Node struct {
	next  *Node
	Value any
}

type SQueue struct {
	currentNode *Node
	anchorNode  *Node

	mutex sync.Mutex
}

func New() *SQueue {
	node := &Node{}
	node.next = node
	return &SQueue{
		currentNode: node,
		anchorNode:  node,
		mutex:       sync.Mutex{},
	}
}

func (q *SQueue) IsEmpty() bool {
	return q.currentNode.next == q.anchorNode
}

func (q *SQueue) Push(val any) {
	node := &Node{
		Value: val,
	}
	node.next = node

	q.mutex.Lock()
	q.currentNode.next.next = node
	q.currentNode = node
	q.mutex.Unlock()
}

func (q *SQueue) PushBack(val any) {
	q.mutex.Lock()
	if q.anchorNode == q.anchorNode.next {
		node := &Node{
			Value: val,
		}
		node.next = node
		q.currentNode.next.next = node
		q.currentNode = node
	} else {
		q.anchorNode.next = &Node{
			Value: val,
			next:  q.anchorNode.next,
		}
	}
	q.mutex.Lock()
}

func (q *SQueue) Pop() (val any, ok bool) {
	q.mutex.Lock()

	node := q.anchorNode.next
	ok = q.anchorNode != node.next
	q.anchorNode.next = node.next
	node.next = q.anchorNode

	q.mutex.Unlock()

	return node.Value, ok
}
