package main

import (
	"fmt"
	"strconv"

	"github.com/shayanh/gogl/lists"
	"github.com/shayanh/gogl/iters"
)

func printList[T any](lst *lists.List[T]) {
	fmt.Print("lists = ")
	iters.ForEach[T](lists.Begin(lst), func(t T) {
		fmt.Print(t, ", ")
	})
	fmt.Println()
}

func main() {
	lst := lists.NewList[int](1, 2, 7)
	lists.PushFront(lst, -9)
	printList(lst)

	lists.Insert(lst, lists.Begin(lst), 11, 12)
	printList(lst)

	lists.Insert(lst, lists.RBegin(lst), 13, 14)
	printList(lst)

	lst2 := lists.NewList[int](iters.Map[int, int](lists.Begin(lst), func(t int) int {
		return t * 2
	})...)
	printList(lst2)

	lst3 := lists.NewList[string](iters.Map[int, string](lists.Begin(lst), func(t int) string {
		return strconv.Itoa(t) + "a"
	})...)
	printList(lst3)
}
