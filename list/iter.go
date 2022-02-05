package list

// FrwIter is a forward iterator.
type FrwIter[T any] struct {
	node *node[T]
	lst  *List[T]
}

func (it *FrwIter[T]) Next() {
	it.node = it.node.next
}

func (it *FrwIter[T]) Prev() {
	it.node = it.node.prev
}

func (it *FrwIter[T]) Done() bool {
	return it.node == it.lst.back
}

func (it *FrwIter[T]) Value() T {
	return it.node.value
}

func (it *FrwIter[T]) SetValue(val T) {
	it.node.value = val
}

// RIter is a reverse iterator.
type RevIter[T any] struct {
	node *node[T]
	lst  *List[T]
}

func (it *RevIter[T]) Next() {
	it.node = it.node.prev
}

func (it *RevIter[T]) Prev() {
	it.node = it.node.next
}

func (it *RevIter[T]) Done() bool {
	return it.node == it.lst.front
}

func (it *RevIter[T]) Value() T {
	return it.node.value
}

func (it *RevIter[T]) SetValue(val T) {
	it.node.value = val
}
