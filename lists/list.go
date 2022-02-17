package lists

import (
	"github.com/shayanh/gogl"
	"github.com/shayanh/gogl/iters"
	"github.com/shayanh/gogl/internal"
)

type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

type List[T any] struct {
	head *node[T]
	tail *node[T]
	size int
}

func NewList[T any](elems ...T) *List[T] {
	head := &node[T]{}
	tail := &node[T]{}

	head.next = tail
	tail.prev = head
	l := &List[T]{
		head: head,
		tail: tail,
	}

	PushBack(l, elems...)
	return l
}

func FromIter[T any](it iters.Iterator[T]) *List[T] {
	l := NewList[T]()
	for it.HasNext() {
		PushBack(l, it.Next())
	}
	return l
}

func Len[T any](l *List[T]) int {
	return l.size
}

// Iter returns a forward iterator to the beginning.
func Iter[T any](l *List[T]) Iterator[T] {
	return &FrwIter[T]{
		node: l.head,
		lst:  l,
	}
}

// RIter returns a reverse iterator to the beginning (in the reverse order).
func RIter[T any](l *List[T]) Iterator[T] {
	return &RevIter[T]{
		node: l.tail,
		lst:  l,
	}
}

func Equal[T comparable](l1, l2 *List[T]) bool {
	if l1.size != l2.size {
		return false
	}
	it1, it2 := Iter(l1), Iter(l2)
	for it1.HasNext() {
		v1 := it1.Next()
		v2 := it2.Next()
		if v1 != v2 {
			return false
		}
	}
	return true
}

func EqualFunc[T any](l1, l2 *List[T], eq gogl.EqualFn[T]) bool {
	if l1.size != l2.size {
		return false
	}
	it1, it2 := Iter(l1), Iter(l2)
	for it1.HasNext() {
		v1 := it1.Next()
		v2 := it2.Next()
		if !eq(v1, v2) {
			return false
		}
	}
	return true
}

func PushBack[T any](l *List[T], elems ...T) {
	Insert(RIter(l), elems...)
}

func PushFront[T any](l *List[T], elems ...T) {
	Insert(Iter(l), elems...)
}

func PopBack[T any](l *List[T]) {
	require(l.size > 0, "list cannot be empty")
	_ = Delete(RIter(l))
}

func PopFront[T any](l *List[T]) {
	require(l.size > 0, "list cannot be empty")
	_ = Delete(Iter(l))
}

func Front[T any](l *List[T]) T {
	require(l.size > 0, "list cannot be empty")
	return l.head.next.value
}

func Back[T any](l *List[T]) T {
	require(l.size > 0, "list cannot be empty")
	return l.tail.prev.value
}

func Reverse[T any](l *List[T]) {
	internal.Reverse[T](Iter(l), RIter(l), l.size)
}

func (l *List[T]) insertBetween(node, prev, next *node[T]) {
	node.next = next
	node.prev = prev

	next.prev = node
	prev.next = node

	l.size += 1
}

// Insert inserts values after the iterator.
func Insert[T any](it Iterator[T], elems ...T) {
	switch typedIt := it.(type) {
	case *FrwIter[T]:
		require(typedIt.Valid(), "iterator must be valid")
		require(typedIt.node.next != nil, "bad iterator")

		for i := len(elems) - 1; i >= 0; i-- {
			elem := elems[i]
			node := &node[T]{value: elem}
			typedIt.lst.insertBetween(node, typedIt.node, typedIt.node.next)
		}
	case *RevIter[T]:
		require(typedIt.Valid(), "iterator must be valid")
		require(typedIt.node.prev != nil, "bad iterator")

		for _, elem := range elems {
			node := &node[T]{value: elem}
			typedIt.lst.insertBetween(node, typedIt.node.prev, typedIt.node)
		}
	default:
		panic("wrong iter type")
	}
}

func (l *List[T]) deleteNode(node *node[T]) (*node[T], *node[T]) {
	prev := node.prev
	next := node.next

	prev.next = next
	next.prev = prev

	node = nil

	l.size -= 1

	return prev, next
}

// Delete deletes the node that the given iterator is pointing to.
// The given iterator will be invalidated and cannot be used. The
// user should use the returned iterator.
func Delete[T any](it Iterator[T]) Iterator[T] {
	switch typedIt := it.(type) {
	case *FrwIter[T]:
		require(typedIt.Valid(), "iterator must be valid")
		require(typedIt.node.prev != nil && typedIt.node.next != nil,
			"bad iterator")

		prev, _ := typedIt.lst.deleteNode(typedIt.node)
		return &FrwIter[T]{
			node: prev,
			lst:  typedIt.lst,
		}
	case *RevIter[T]:
		require(typedIt.Valid(), "iterator must be valid")
		require(typedIt.node.prev != nil && typedIt.node.next != nil,
			"bad iterator")

		_, next := typedIt.lst.deleteNode(typedIt.node)
		return &RevIter[T]{
			node: next,
			lst:  typedIt.lst,
		}
	default:
		panic("wrong iter type")
	}
}
