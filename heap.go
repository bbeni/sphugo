/* Heap methods
Author: Benjamin Frölich

This packacke should provide methods for heapification of arrays.

TODOs:
	Goals:
		- Min-Heap
		- Max-Heap
		- Maybe Fibonacci Heap
		- Maybe Binomial Heap
		- ...

	Functionality:
		- BuildHeap
		- DecreaseKey
		- Insert
		- Remove
		- Find Min/Max
		- Extract Min/Max
*/

package main

import (
	"fmt"
	"math/rand"
	"cmp"
)

/* Heap indexed in the following way:
	i parent
		-> (i+1)*2-1 left  child
		-> (i+1)*2   right child */

const ARRAY_CAP = 1024 // Dynamic array initial capacity

/* Heapify()

remakes it a heap if all childs left and right of array[i] fulfill
min heap condition!

min heap condition:
	parent <= left  child and
 	parent <= right child

tail recursive so we are using loop instead of recursive solution */

func Heapify[T cmp.Ordered](array []T, i int) {
	for {
		l := (i + 1) * 2 - 1
		r := (i + 1) * 2
		min_index := i

		if l < len(array) && array[l] < array[min_index] {
			min_index = l
		}

		if r < len(array) && array[r] < array[min_index] {
			min_index = r
		}

		if min_index == i {
			break
		}

		array[min_index], array[i] = array[i], array[min_index]
		i = min_index
	}
}

/* Recursive version of Heapify() */

func HeapifyRec[T cmp.Ordered](array []T, i int) {

	l := (i + 1) * 2 - 1
	r := (i + 1) * 2
	min_index := i

	if l < len(array) && array[l] < array[min_index] {
		min_index = l
	}

	if r < len(array) && array[r] < array[min_index] {
		min_index = r
	}

	if min_index != i {
		array[min_index], array[i] = array[i], array[min_index]
		HeapifyRec(array, min_index)
	}
}

func BuildHeap[T cmp.Ordered](array []T) {

	if len(array) < 2 {
		return
	}

	for i := len(array)/2 - 1; i >= 0; i-- {
		Heapify(array, i)
	}
}

func dump[T cmp.Ordered](x []T, index, level int) {

	if index >= len(x) { return }

	l := (index + 1)*2 - 1
	r := (index + 1)*2

	if r < len(x){
		for range level { fmt.Print("    ") }
		fmt.Printf("R=%v\n", x[r])
	}

	if l < len(x){
		for range level { fmt.Print("    ") }
		fmt.Printf("L=%v\n", x[l])
	}

	if r < len(x) { dump(x, r, level+1) }
	if l < len(x) { dump(x, l, level+1) }
}

// print the Heap in [node - right - left] top down fashin

func DumpHeap1[T cmp.Ordered](array []T) {

	fmt.Printf("H=%v\n", array[0])
	dump(array, 0, 1)

}

// print the Heap like in text books, but we are in a terminal

func DumpHeap[T cmp.Ordered](array []T) {
	const spacing = " "
	level := 0
	for (len(array) >> level) != 0 {
		level++
	}

	fmt.Println()
	for range 1 << level - 6 { fmt.Printf(spacing) }
	fmt.Println("[Root Node]")
	for i := range level {
		offset := 1 << i - 1
		for j := range offset + 1 {
			index := j + offset
			if index < len(array) {
				for range 1 << (level - i) - 1 { fmt.Printf(spacing) }
				fmt.Printf("%v", array[index])
				for range 1 << (level - i) - 1 { fmt.Printf(spacing) }
			}
		}
		fmt.Print("\n\n")
	}
	fmt.Println()
}

func main() {

	rand.Seed(101)

	array := make([]int, 0, ARRAY_CAP)

	for _ = range 26 {
		array = append(array, rand.Int() % 101)
	}

	fmt.Printf("\n")
	fmt.Printf("initial array was:    %v\n", array)

	BuildHeap(array)

	fmt.Printf("after heapification:  %v\n", array)
	fmt.Println("visualisation to check for correctness:")

	DumpHeap(array)

}