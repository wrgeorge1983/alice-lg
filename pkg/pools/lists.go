package pools

import (
	"sync"
)

// Node is a node in the tree.
type Node struct {
	children []*Node
	value    int
	counter  int
	ptr      interface{}
}

// Internally acquire list by traversing the tree and
// creating nodes if required.
func (n *Node) traverse(list interface{}, tail []int) interface{} {
	head := tail[0]
	tail = tail[1:]

	// Seek for value in children
	var child *Node
	for _, c := range n.children {
		if c.value == head {
			child = c
		}
	}
	if child == nil {
		// Insert child
		child = &Node{
			children: []*Node{},
			value:    head,
			ptr:      nil,
		}
		n.children = append(n.children, child)
	}

	// Set list ptr if required
	if len(tail) == 0 {
		if child.ptr == nil {
			child.ptr = list
		}
		return child.ptr
	}

	return child.traverse(list, tail)
}

// A IntList pool can be used to deduplicate
// lists of integers. Like an AS path.
//
// A Tree datastructure is used.
type IntList struct {
	root *Node
	sync.Mutex
}

// NewIntList creates a new int list pool
func NewIntList() *IntList {
	return &IntList{
		root: &Node{
			ptr: []int{},
		},
	}
}

// Acquire int list from pool
func (p *IntList) Acquire(list []int) []int {
	p.Lock()
	defer p.Unlock()

	if len(list) == 0 {
		return p.root.ptr.([]int) // root
	}
	return p.root.traverse(list, list).([]int)
}

// A StringList pool can be used for deduplicating lists
// of strings. (This is a variant of an int list, as string
// values are converted to int.
type StringList struct {
	root   *Node
	values map[string]int
	head   int
	sync.Mutex
}

// NewStringList creates a new string list.
func NewStringList() *StringList {
	return &StringList{
		head:   1,
		values: map[string]int{},
		root: &Node{
			ptr: []string{},
		},
	}
}

// Acquire the string list pointer from the pool
func (p *StringList) Acquire(list []string) []string {
	if len(list) == 0 {
		return p.root.ptr.([]string) // root
	}

	// Make idenfier list
	id := make([]int, len(list))
	for i, s := range list {
		// Resolve string value into int
		v, ok := p.values[s]
		if !ok {
			p.head++
			p.values[s] = p.head
			v = p.head
		}
		id[i] = v
	}

	return p.root.traverse(list, id).([]string)
}