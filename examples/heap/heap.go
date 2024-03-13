/* Heap Example

Author: Benjamin Frölich

Heap example. for more info and implementation details see tg/heap.go

*/

package main

import (
	"fmt"
	"math/rand"
	"github.com/bbeni/treego/tg"
)

/* Heap indexed in the following way:
	i parent
		-> (i+1)*2-1 left  child
		-> (i+1)*2   right child
*/

const ARRAY_CAP = 1024 // Dynamic array initial capacity

func main() {

	rand.Seed(101)

	array := make([]int, 0, ARRAY_CAP)

	for _ = range 26 {
		array = append(array, rand.Int() % 90 + 9)
	}

	fmt.Printf("\n")
	fmt.Printf("initial array was:    %v\n", array)

	tg.BuildHeap(array)

	fmt.Printf("after heapification:  %v\n\n", array)
	fmt.Println("visualisation to check for correctness:")
	tg.DumpHeap(array)

	fmt.Println("Insert 0 into the Heap:")
	array = tg.Insert(array, 0)
	tg.DumpHeap(array)

	array, _, _ = tg.ExtractMin(array)
	fmt.Print("ExtractMin from Heap:\n\n")
	tg.DumpHeap(array)

	array, _, _ = tg.Replace(array, 33)
	fmt.Print("Replace with 33 with root node:\n\n")
	tg.DumpHeap(array)

}