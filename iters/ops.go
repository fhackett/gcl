package iters

import (
	"golang.org/x/exp/constraints"

	"github.com/shayanh/gcl"
)

// Equal determines if the elements of two iterators are equal. It returns true
// if all elements are equal and both iterators have the same number of
// elements. If the first iterator has l1 elements and the second iterator has
// l2 elements, Equal advances each iterator min(l1, l2) steps.
func Equal[T comparable](it1, it2 Iterator[T]) bool {
	return ForAll(Map(Zip(it1, it2), func(pair gcl.Pair[T, T]) bool {
		return pair.First == pair.Second
	}))
}

func ForAll(it Iterator[bool]) bool {
	result := true
	ForEach(it, func(elem bool) {
		result = result && elem
	})
	return result
}

func Exists(it Iterator[bool]) bool {
	result := false
	ForEach(it, func(elem bool) {
		result = result || elem
	})
	return result
}

// EqualFunc works the same as Equal, but it uses the function eq for element
// comparison.
func EqualFunc[T1 any, T2 any](it1 Iterator[T1], it2 Iterator[T2], eq gcl.EqualFn[T1, T2]) bool {
	return ForAll(Map(Zip(it1, it2), func(pair gcl.Pair[T1, T2]) bool {
		return eq(pair.First, pair.Second)
	}))
}

// Reduce applies a function of two arguments cumulatively to the items of
// the given iterator from the beginning to the end.
func Reduce[T any](it Iterator[T], fn func(T, T) T) (acc T, ok bool) {
	ForEach(it, func(elem T) {
		if !ok {
			ok = true
			acc = elem
		} else {
			acc = fn(acc, elem)
		}
	})
	return
}

// Fold applies a function of two arguments cumulatively to the items of
// the given iterator from the beginning to the end. Fold gives an initial
// value and start its operation by using the initial value.
// Fold moves the given iterator it to its end such that after a Fold
// call it.HasNext() will be false.
func Fold[T any, V any](it Iterator[T], fn func(V, T) V, init V) (acc V) {
	acc = init
	ForEach(it, func(elem T) {
		acc = fn(acc, elem)
	})
	return
}

// Max returns the maximum element in an iterator of any ordered type.
func Max[T constraints.Ordered](it Iterator[T]) (max T, ok bool) {
	return Reduce(it, func(left, right T) T {
		if left < right {
			return right
		} else {
			return left
		}
	})
}

// MaxFunc returns the maximum element in an iterator and uses the given
// less function for comparison.
func MaxFunc[T any](it Iterator[T], less gcl.LessFn[T]) (max T, ok bool) {
	return Reduce(it, func(left, right T) T {
		if less(left, right) {
			return right
		} else {
			return left
		}
	})
}

// Min returns the minimum element in an iterator of any ordered type.
func Min[T constraints.Ordered](it Iterator[T]) (min T, ok bool) {
	return Reduce(it, func(left, right T) T {
		if left < right {
			return left
		} else {
			return right
		}
	})
}

// MinFunc returns the minimum element in an iterator and uses the given
// less function for comparison.
func MinFunc[T any](it Iterator[T], less gcl.LessFn[T]) (min T, ok bool) {
	return Reduce(it, func(left, right T) T {
		if less(left, right) {
			return left
		} else {
			return right
		}
	})
}

// Sum returns sum of the elements in an iterator of any numeric type.
// Sum moves the given iterator it to its end such that after a Sum
// call it.HasNext() will be false.
func Sum[T gcl.Number](it Iterator[T]) (sum T) {
	sum = 0
	ForEach(it, func(elem T) {
		sum += elem
	})
	return
}

// Find returns the first element in an iterator that satisfies pred. The
// returned boolean value indicates if such an element exists. If a satisfying
// element exists, the given iterator it advances the first satisfying element
// by one step, otherwise Find moves the iterator it to its end.
func Find[T any](it Iterator[T], pred func(T) bool) (value T, ok bool) {
	it.Generate(IterConfig{}, func(elem T) IterState {
		if pred(elem) {
			value, ok = elem, true
			return IterStop
		}
		return IterOk
	})
	ok = false
	return
}

// ForEach calls a function on each element of an iterator.
func ForEach[T any](it Iterator[T], fn func(T)) {
	it.Generate(IterConfig{}, func(elem T) IterState {
		fn(elem)
		return IterOk
	})
}
