# ESC 202 Assignments

My attempt at the assignments at ESC 202 course at UZH written in [GO](https://go.dev/ "Go Language"). 🦆

## 2d Particles Binary Partition

partitions 2200 uniformely distributed particles in a 2d space and generates tree.png.

```console
go run tree-partition-2d.go
```

The function Partition() partitions an array of type Particle based on their 2d position. They are compared to a pivot value called middle in a "bubble sort like" manner in a specified axis that can either be "Vertical" or "Horizontal". The tests should cover most edge cases. Returns two partitioned slices a, b (just indices of array in Go).

The function Treebuild() recurses and partitions an array of N_PARTICLES length int Cells that have maximally MAX_PARTICLES_PER_CELL particles. The SPLIT_FRACTION determines the fraction of space in the specific direction for left/total or top/total.

### Visualisation

A png picture is generated from a tree with the make_tree_png() function. The following parameters are used for generating the picture:

	N_PARTICLES = 2200
	MAX_PARTICLES_PER_CELL = 8
	SPLIT_FRACTION = 0.5
	IMAGE_W = 512*2
	IMAGE_H = 512*2

#### Treebuild visualization

![](tree.png)
