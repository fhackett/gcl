package hmaps

import (
	"github.com/shayanh/gcl"
	"github.com/shayanh/gcl/gomaps"
	"github.com/shayanh/gcl/iters"
	"github.com/shayanh/gcl/lists"
)

type Map[K, V any] struct {
	hashFn    gcl.HashFn[K]
	compareFn gcl.CompareFn[K, K]
	impl      map[uint64]*lists.List[gcl.MapElem[K, V]]
	len       int
}

func New[K, V any](hashFn gcl.HashFn[K], compareFn gcl.CompareFn[K, K]) *Map[K, V] {
	return &Map[K, V]{
		hashFn:    hashFn,
		compareFn: compareFn,
		impl:      make(map[uint64]*lists.List[gcl.MapElem[K, V]]),
	}
}

func (mp *Map[K, V]) Put(key K, value V) {
	h := mp.hashFn(key)
	if buckets, ok := mp.impl[h]; ok {
		it := lists.IterMut(buckets)
		for it.HasNext() {
			bucket := it.Next()
			if mp.compareFn(bucket.Key, key) == 0 {
				bucket.Value = value
				return
			}
		}
		it.Insert(gcl.MapElem[K, V]{
			Key:   key,
			Value: value,
		})
		mp.len++
	} else {
		mp.impl[h] = lists.New(gcl.MapElem[K, V]{
			Key:   key,
			Value: value,
		})
		mp.len++
	}
}

func (mp *Map[K, V]) Get(key K) (V, bool) {
	h := mp.hashFn(key)
	if buckets, ok := mp.impl[h]; ok {
		it := lists.Iter(buckets)
		for it.HasNext() {
			bucket := it.Next()
			if mp.compareFn(bucket.Key, key) == 0 {
				return bucket.Value, true
			}
		}
	}
	var result V
	return result, false
}

func (mp *Map[K, V]) Delete(key K) {
	h := mp.hashFn(key)
	if buckets, ok := mp.impl[h]; ok {
		it := lists.IterMut(buckets)
		for it.HasNext() {
			bucket := it.Next()
			if mp.compareFn(bucket.Key, key) == 0 {
				it.Delete()
				mp.len--
				return
			}
		}
	}
}

func (mp *Map[K, V]) Len() int {
	return mp.len
}

func Iter[K, V any](mp *Map[K, V]) *iters.FlatMapIter[gcl.MapElem[uint64, *lists.List[gcl.MapElem[K, V]]], gcl.MapElem[K, V], *gomaps.Iterator[uint64, *lists.List[gcl.MapElem[K, V]]], *lists.FrwIter[gcl.MapElem[K, V]]] {
	innerIter := gomaps.Iter(mp.impl)
	return iters.FlatMap[gcl.MapElem[uint64, *lists.List[gcl.MapElem[K, V]]], gcl.MapElem[K, V], *gomaps.Iterator[uint64, *lists.List[gcl.MapElem[K, V]]], *lists.FrwIter[gcl.MapElem[K, V]]](innerIter, func(pair gcl.MapElem[uint64, *lists.List[gcl.MapElem[K, V]]]) *lists.FrwIter[gcl.MapElem[K, V]] {
		return lists.Iter(pair.Value)
	})
}

func (mp *Map[K, V]) Equal(other *Map[K, V], compareFn gcl.CompareFn[V, V]) bool {
	if mp.Len() != other.Len() {
		return false
	}
	it := Iter(mp)
	for it.HasNext() {
		pair := it.Next()
		if value, ok := other.Get(pair.Key); ok {
			if compareFn(pair.Value, value) != 0 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
