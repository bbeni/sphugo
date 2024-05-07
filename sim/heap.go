/*
	Heap methods

Author: Benjamin Frölich

This packacke should provide methods for heapification of arrays.

TODOs:

	Goals:
		- [x] Min-Heap
		- [ ] Max-Heap
		- Maybe Fibonacci Heap
		- Maybe Binomial Heap
		- ...

	Functionality:
		- [x] BuildHeap
		- [ ] DecreaseKey
		- [x] Insert
		- [ ] Remove
		- [x] Replace
		- [x] Find Min/Max
		- [x] Extract Min/Max
*/
package sim

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
)

/* Heap indexed in the following way:
i parent
	-> i*2 + 1   left  child
	-> i*2 + 2   right child */

/* Heapify()

remakes it a heap if all childs left and right of array[i] fulfill
min heap condition!

min heap condition:
	parent <= left  child and
 	parent <= right child

tail recursive so we are using loop instead of recursive solution */

func Heapify[T cmp.Ordered](array []T, i int) {
	for {
		l := i*2 + 1
		r := i*2 + 2
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

	l := (i+1)*2 - 1
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

func Insert[T cmp.Ordered](array []T, element T) []T {
	array = append(array, element)
	index := len(array) - 1
	parent := len(array)/2 - 1
	for parent >= 0 && array[index] < array[parent] {
		array[parent], array[index] = array[index], array[parent]
		index, parent = parent, (parent+1)/2-1
	}
	return array
}

func FindMin[T cmp.Ordered](array []T) (T, error) {
	if len(array) == 0 {
		var def T
		return def, errors.New("in FindMin(): array can not be empty")
	}
	return array[0], nil
}

func ExtractMin[T cmp.Ordered](array []T) ([]T, T, error) {
	if len(array) == 0 {
		var def T
		return array, def, errors.New("in ExtractMin(): array can not be empty")
	}
	min := array[0]
	array[0] = array[len(array)-1]
	array = array[:len(array)-1]
	Heapify(array, 0)
	return array, min, nil
}

func Replace[T cmp.Ordered](array []T, element T) ([]T, T, error) {
	if len(array) == 0 {
		var def T
		return array, def, errors.New("in Replace(): array can not be empty")
	}

	min := array[0]
	array[0] = element
	Heapify(array, 0)
	return array, min, nil
}

func dump[T cmp.Ordered](x []T, index, level int) {

	if index >= len(x) {
		return
	}

	l := (index+1)*2 - 1
	r := (index + 1) * 2

	if r < len(x) {
		for range level {
			fmt.Print("    ")
		}
		fmt.Printf("R=%v\n", x[r])
	}

	if l < len(x) {
		for range level {
			fmt.Print("    ")
		}
		fmt.Printf("L=%v\n", x[l])
	}

	if r < len(x) {
		dump(x, r, level+1)
	}
	if l < len(x) {
		dump(x, l, level+1)
	}
}

// print the Heap in [node - right - left] top down fashin

func DumpHeap1[T cmp.Ordered](array []T) {

	fmt.Printf("H=%v\n", array[0])
	dump(array, 0, 1)

}

// print the Heap like in text books, but we are in a terminal

func DumpHeap[T cmp.Ordered](array []T) {

	spaces := func(amount int) string {
		return strings.Repeat(" ", amount)
	}

	level := 0
	for (len(array) >> level) != 0 {
		level++
	}

	fmt.Println()
	fmt.Println(spaces(1<<level-6) + "[Root Node]")
	for i := range level {
		offset := 1<<i - 1
		for j := range offset + 1 {
			index := j + offset
			if index < len(array) {
				n_spaces := 1<<(level-i) - 1
				fmt.Printf(spaces(n_spaces)+"%v"+spaces(n_spaces), array[index])
			}
		}
		fmt.Print("\n\n")
		mid_count := 1<<(level-i) - 2
		mid_spaces := spaces(mid_count)
		for j := range offset + 1 {
			index := j + offset
			if index < len(array)/2 {
				n_spaces := 1 << (level - i - 1)
				fmt.Printf(spaces(n_spaces) + "/")
				if index <= len(array)/2-2 || len(array)%2 == 1 {
					fmt.Printf(mid_spaces + "\\" + spaces(n_spaces))
				}
			}
		}
		fmt.Print("\n")
	}
}

func QuickSort(particles []*Particle, cmpValue func(*Particle) int) {
	if len(particles) > 0 {
		left, right := partitionQS(particles, cmpValue)
		QuickSort(left, cmpValue)
		QuickSort(right, cmpValue)
	}
}

func partitionQS(ps []*Particle, cmpValue func(*Particle) int) ([]*Particle, []*Particle) {

	i := 0
	j := len(ps) - 2

	if len(ps) == 0 {
		panic("QS unreachable")
	}

	pivot := cmpValue(ps[len(ps)-1])

	for i < j {
		for i < j && cmpValue(ps[i]) <= pivot {
			i++
		}
		for i < j && cmpValue(ps[j]) > pivot {
			j--
		}

		if cmpValue(ps[i]) > cmpValue(ps[j]) {
			ps[i], ps[j] = ps[j], ps[i]
		}
	}

	if cmpValue(ps[i]) > pivot {
		ps[i], ps[len(ps)-1] = ps[len(ps)-1], ps[i]
	} else {
		i = len(ps) - 1
	}

	return ps[:i], ps[i+1:]
}
