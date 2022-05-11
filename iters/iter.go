// Package iters defines the general iterator interface and provides different
// operations on top of them.
package iters

import "github.com/shayanh/gcl"

type IterConfig struct {
	ShouldStop bool
	Defers     []func()
}

func (config IterConfig) IsStop() bool {
	return config.ShouldStop
}

func (config IterConfig) AddDefers(defers []func()) IterConfig {
	config.Defers = append(config.Defers, defers...)
	return config
}

func (config IterConfig) WithoutDefers() IterConfig {
	config.Defers = nil
	return config
}

func (config IterConfig) ToInitState() IterState {
	if config.ShouldStop {
		return IterStop
	} else {
		return IterOk
	}
}

func (config IterConfig) RunDefers() {
	for i := len(config.Defers) - 1; i >= 0; i-- {
		config.Defers[i]()
	}
}

type IterState int

func (state IterState) IsStop() bool {
	return state == IterStop
}

func (state IterState) ToConfig() IterConfig {
	return IterConfig{ShouldStop: state == IterStop}
}

func (state IterState) Concat(other IterState) IterState {
	switch state {
	case IterOk:
		return other
	case IterStop:
		return IterStop
	}
	panic("unreachable")
}

const (
	IterOk IterState = iota
	IterStop
)

type IterFunc[T any] func(elem T) IterState

type Iterator[T any] interface {
	Generate(config IterConfig, fn IterFunc[T])
}

type skipIterator[T any] struct {
	super Iterator[T]
	count int
}

func (it skipIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	countToSkip := it.count
	it.super.Generate(config, func(elem T) IterState {
		if countToSkip > 0 {
			countToSkip--
			return IterOk
		}
		return fn(elem)
	})
}

func Skip[T any](it Iterator[T], count int) Iterator[T] {
	return skipIterator[T]{
		super: it,
		count: count,
	}
}

type takeIterator[T any] struct {
	super Iterator[T]
	count int
}

func (it takeIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	countRemaining := it.count
	it.super.Generate(config, func(elem T) IterState {
		if countRemaining == 0 {
			return IterStop
		}
		countRemaining--
		return fn(elem)
	})
}

func Take[T any](it Iterator[T], count int) Iterator[T] {
	return takeIterator[T]{
		super: it,
		count: count,
	}
}

type mapIterator[T, U any] struct {
	super Iterator[T]
	fn    func(T) U
}

func (it mapIterator[T, U]) Generate(config IterConfig, fn IterFunc[U]) {
	it.super.Generate(config, func(elem T) IterState {
		return fn(it.fn(elem))
	})
}

// Map applies the function fn on elements the given iterator it and returns a
// new iterator over the mapped values. Map moves the given iterator it to its
// end such that after a Map call it.HasNext() will be false.
// Map is lazy, in a way that if you don't consume the returned iterator
// nothing will happen.
func Map[T, U any](it Iterator[T], fn func(T) U) Iterator[U] {
	return mapIterator[T, U]{
		super: it,
		fn:    fn,
	}
}

type filterIterator[T any] struct {
	super Iterator[T]
	pred  func(T) bool
}

func (it filterIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	it.super.Generate(config, func(elem T) IterState {
		if it.pred(elem) {
			return fn(elem)
		} else {
			return IterOk
		}
	})
}

// Filter filters elements of an iterator that satisfy the pred.
// Filter returns an iterator over the filtered elements. Filter moves the given
// iterator it to its end such that after a Filter call it.HasNext() will be
// false.
// Filter is lazy, in a way that if you don't consume the returned iterator
// nothing will happen.
func Filter[T any](it Iterator[T], pred func(T) bool) Iterator[T] {
	return filterIterator[T]{
		super: it,
		pred:  pred,
	}
}

type flatMapIterator[T, U any] struct {
	super Iterator[T]
	fn    func(T) Iterator[U]
}

func (it flatMapIterator[T, U]) Generate(config IterConfig, fn IterFunc[U]) {
	it.super.Generate(config, func(elem T) IterState {
		nestedIter := it.fn(elem)
		state := IterOk
		nestedIter.Generate(IterConfig{}, func(elem U) IterState {
			state = fn(elem)
			return state
		})
		return state
	})
}

func FlatMap[T, U any](it Iterator[T], fn func(T) Iterator[U]) Iterator[U] {
	return flatMapIterator[T, U]{
		super: it,
		fn:    fn,
	}
}

type deferIterator[T any] struct {
	super  Iterator[T]
	defers []func()
}

func (it deferIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	it.super.Generate(config.AddDefers(it.defers), fn)
}

func Defer[T any](it Iterator[T], defers ...func()) Iterator[T] {
	return deferIterator[T]{
		super:  it,
		defers: defers,
	}
}

type repeatIterator[T any] struct {
	generator func() T
}

func (it repeatIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	state := config.ToInitState()
	for !state.IsStop() {
		elem := it.generator()
		state = state.Concat(fn(elem))
	}
	config.RunDefers()
}

func Repeat[T any](generator func() T) Iterator[T] {
	return repeatIterator[T]{
		generator: generator,
	}
}

type zipIterator[T, U any] struct {
	left  Iterator[T]
	right Iterator[U]
}

func (it zipIterator[T, U]) Generate(config IterConfig, fn IterFunc[gcl.Pair[T, U]]) {
	iterStateCh := make(chan IterState)
	rightElems := make(chan U)
	go func() {
		it.right.Generate(config.WithoutDefers(), func(elem U) IterState {
			rightElems <- elem
			return <-iterStateCh
		})
		close(rightElems)
	}()
	it.left.Generate(config.WithoutDefers(), func(leftElem T) IterState {
		rightElem, ok := <-rightElems
		if !ok {
			return IterStop
		}
		iterState := fn(gcl.Pair[T, U]{
			First:  leftElem,
			Second: rightElem,
		})
		iterStateCh <- iterState
		return iterState
	})

	config.RunDefers()
}

// Zip zips the two given iterators and returns a single iterator over
// gcl.Pair values.
func Zip[T, U any](left Iterator[T], right Iterator[U]) Iterator[gcl.Pair[T, U]] {
	return zipIterator[T, U]{
		left:  left,
		right: right,
	}
}

type ConcatIterator[T any] struct {
	first, second Iterator[T]
}

func (it ConcatIterator[T]) Generate(config IterConfig, fn IterFunc[T]) {
	var state = config.ToInitState()
	it.first.Generate(config.WithoutDefers(), func(elem T) IterState {
		state = fn(elem)
		return state
	})
	it.second.Generate(state.ToConfig(), func(elem T) IterState {
		return fn(elem)
	})
	config.RunDefers()
}

func Concat[T any](first, second Iterator[T]) Iterator[T] {
	return ConcatIterator[T]{
		first:  first,
		second: second,
	}
}
