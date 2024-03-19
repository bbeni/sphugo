# ESC 202 Simulations

My attempt at the assignments at ESC 202 course at UZH written in [GO](https://go.dev/ "Go Language"). 🦆 The Goal seems to be to have a working Smooth Particle Hydrodynamics code. Go as a language was chosen to easily parallelize the simulation and have fast compiled code.

## 1. Example - Binary Partition 2d Particles

Goal:

>Build a binary tree of cells using the partitioning of particles function we introduced in class. The hard part is making sure your partition function is really bomb proof, check all “edge cases” (e.g., no particles in the cell, all particles on one side or other of the partition, already partitioned data, particles in the inverted order of the partition, etc…). Write a boolean test functions for each of these cases. Call all test functions in sequence and check if they all succeed. Once you have this, then recursively partition the partitions and build cells linked into a tree as you go. Partition alternately in x and y dimensions, or simply partition the longest dimension of the given cell.

partitions 2200 uniformely distributed particles in a 2d space and generates tree.png.

```console
go run ./examples/tree-partition/
```

The function Partition() partitions an array of type Particle based on their 2d position. They are compared to a pivot value called middle in a "bubble sort like" manner in a specified axis that can either be "Vertical" or "Horizontal". Returns two partitioned slices a, b. The tests should cover most edge cases (run ```go test -v ./tg```).

The function Treebuild() recurses and partitions an array of N_PARTICLES length int Cells that have maximally MAX_PARTICLES_PER_CELL particles. The SPLIT_FRACTION determines the fraction of space in the specific direction for left/total or top/total.

### Visualisation

A png picture is generated from a tree with the MakeTreePng() function. The following parameters are used for generating the picture:

	N_PARTICLES = 2200
	MAX_PARTICLES_PER_CELL = 8
	SPLIT_FRACTION = 0.5
	IMAGE_W = 512*2
	IMAGE_H = 512*2

#### Treebuild visualization

![](tree.png)

## 2. Example - Heap Implementation

Showcase BuildHeap, Insert, ExtractMin, Replace functionality. Dumptree function used for visualizing the tree in terminal (text form) and check for correctness.

```console
go run ./examples/heap/
```

Output of this example:

```console

initial array was:    [31 37 82 83 33 54 39 42 62 49 84 59 88 26 27 21 92 97 87 49 33 9 42 49 88 67]
after heapification:  [9 21 26 31 33 49 27 42 62 37 33 54 67 39 82 83 92 97 87 49 49 84 42 59 88 88]

visualisation to check for correctness:

                          [Root Node]
                               9                               

                /                              \                
               21                              26               

        /              \                /              \        
       31              33              49              27       

    /      \        /      \        /      \        /      \    
   42      62      37      33      54      67      39      82   

  /  \    /  \    /  \    /  \    /  \    /
 83  92  97  87  49  49  84  42  59  88  88 


Insert 0 into the Heap:

                          [Root Node]
                               0                               

                /                              \                
               21                              9               

        /              \                /              \        
       31              33              26              27       

    /      \        /      \        /      \        /      \    
   42      62      37      33      54      49      39      82   

  /  \    /  \    /  \    /  \    /  \    /  \  
 83  92  97  87  49  49  84  42  59  88  88  67 


ExtractMin from Heap:


                          [Root Node]
                               9                               

                /                              \                
               21                              26               

        /              \                /              \        
       31              33              49              27       

    /      \        /      \        /      \        /      \    
   42      62      37      33      54      67      39      82   

  /  \    /  \    /  \    /  \    /  \    /
 83  92  97  87  49  49  84  42  59  88  88 


Replace with 33 with root node:


                          [Root Node]
                               21                               

                /                              \                
               31                              26               

        /              \                /              \        
       33              33              49              27       

    /      \        /      \        /      \        /      \    
   42      62      37      33      54      67      39      82   

  /  \    /  \    /  \    /  \    /  \    /
 83  92  97  87  49  49  84  42  59  88  88 

```

## 3. Nearest Neigbours

Goal:

>Implement the k nearest neighbor search. Use the priority queue given in the Python template and implement “replace” and “key” functions. Use the particle to cell distance function from the lecture notes or the celldist2() given in the Python template. Are they the same? Optional: Also implement the ball search algorithm given in the lecture notes.


The function `FindNearestNeighbours()` acts on one Particle and uses a Prority Queue, implemented similarly to the Heap shown before, to find the lowest distance Neighbours. NN_SIZE=32 constant defines the nearest neighbour count.

```console
go run ./examples/nearest-neighbours/
```

It generates two images with 220 particles. The first shows the non periodic version of the particle.FindNearestNeighbour function. The tree cells are also shown in red.

![](nearest_neighbours.png)

The periodic visualization inculeds a the bounding 'spheres' of each tree cell instead of the tree cell.

![](nearest_neighbours_periodic.png)

## Tests

To run all tests (Partition(), BoundingSpheres() covered for now):

```console
go test -v ./...
```
