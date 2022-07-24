// package toroidal contains types and functions related to toroidal quadtree data structures.
package toroquatre

import (
	"fmt"
	"sort"
	"strings"
)

type Vector struct {
	X, Y float64
}

type ToroidalQuadtree struct {
	nodeLimit int
	leafMap   map[uint64]Vector
	root      *tree
}

type tree struct {
	tqt *ToroidalQuadtree
	// Bounding box for the tree.
	topLeft     Vector
	bottomRight Vector
	// If the tree has nodeLimit or fewer leaves, their IDs are here.
	leaves []uint64
	// Once the tree has more than nodeLimit leaves, it instead contains 4 subtrees.
	// They are in quadrant order: Northwest, Northeast, Southwest, Southeast.
	branches [4]*tree
	// The total number of items within this part of the tree
	count int
}

func medianLocation(ids []uint64, leafMap map[uint64]Vector) Vector {
	sortedX := make([]float64, len(ids))
	sortedY := make([]float64, len(ids))
	for i, id := range ids {
		sortedX[i] = leafMap[id].X
		sortedY[i] = leafMap[id].Y
	}
	sort.Float64s(sortedX)
	sort.Float64s(sortedY)

	var result Vector
	// Even length list: take the average of the two middle elements
	if len(ids)%2 == 0 {
		idx1 := len(ids) / 2
		idx2 := idx1 - 1
		result.X = (sortedX[idx1] + sortedX[idx2]) / 2
		result.Y = (sortedY[idx1] + sortedY[idx2]) / 2
		// Odd length list: take the middle element
	} else {
		idx := len(ids) / 2
		result.X = sortedX[idx]
		result.Y = sortedY[idx]
	}

	return result
}

func (t *tree) subdivide() {
	// Subdivide on the median location within the region
	median := medianLocation(t.leaves, t.tqt.leafMap)

	// Set up the subregions
	t.branches[0] = &tree{
		tqt:         t.tqt,
		topLeft:     t.topLeft,
		bottomRight: median,
		leaves:      make([]uint64, 0, t.tqt.nodeLimit),
	}
	t.branches[1] = &tree{
		tqt:         t.tqt,
		topLeft:     Vector{X: median.X, Y: t.topLeft.Y},
		bottomRight: Vector{X: t.bottomRight.X, Y: median.Y},
		leaves:      make([]uint64, 0, t.tqt.nodeLimit),
	}
	t.branches[2] = &tree{
		tqt:         t.tqt,
		topLeft:     Vector{X: t.topLeft.X, Y: median.Y},
		bottomRight: Vector{X: median.X, Y: t.bottomRight.Y},
		leaves:      make([]uint64, 0, t.tqt.nodeLimit),
	}
	t.branches[3] = &tree{
		tqt:         t.tqt,
		topLeft:     median,
		bottomRight: t.bottomRight,
		leaves:      make([]uint64, 0, t.tqt.nodeLimit),
	}

	// Move all the elements into their proper region
	for _, id := range t.leaves {
		loc := t.tqt.leafMap[id]
		var destRegion int
		switch {
		case loc.X < median.X && loc.Y < median.Y:
			destRegion = 0
		case loc.X >= median.X && loc.Y < median.Y:
			destRegion = 1
		case loc.X < median.X && loc.Y >= median.Y:
			destRegion = 2
		default:
			destRegion = 3
		}
		t.branches[destRegion].leaves = append(t.branches[destRegion].leaves, id)
		t.branches[destRegion].count++
	}
	t.leaves = nil
}

func (t *tree) put(id uint64, p Vector) {
	if t.branches[0] == nil {
		t.leaves = append(t.leaves, id)
		if len(t.leaves) > t.tqt.nodeLimit {
			t.subdivide()
		}
	} else {
		median := t.branches[0].bottomRight
		var destRegion int
		switch {
		case p.X < median.X && p.Y < median.Y:
			destRegion = 0
		case p.X >= median.X && p.Y < median.Y:
			destRegion = 1
		case p.X < median.X && p.Y >= median.Y:
			destRegion = 2
		default:
			destRegion = 3
		}
		t.branches[destRegion].put(id, p)
	}
	t.count++
}

func (t *tree) remove(id uint64, leafMap map[uint64]Vector) {
	if len(t.leaves) != 0 {
		// Find the element to remove, overwrite it with the last element, and subslice the leaves
		for i := 0; i < len(t.leaves)-1; i++ {
			if t.leaves[i] == id {
				t.leaves[i] = t.leaves[len(t.leaves)-1]
				break
			}
		}
		t.leaves = t.leaves[:len(t.leaves)-1]
	} else {
		loc := leafMap[id]
		// Delete from the proper subtree
		median := t.branches[0].bottomRight
		var destRegion int
		switch {
		case loc.X < median.X && loc.Y < median.Y:
			destRegion = 0
		case loc.X >= median.X && loc.Y < median.Y:
			destRegion = 1
		case loc.X < median.X && loc.Y >= median.Y:
			destRegion = 2
		default:
			destRegion = 3
		}
		t.branches[destRegion].remove(id, leafMap)
	}
	t.count--

	// N.B.: Do this after recursing on the child region so that we know it has been flattened
	// before we merge it here.
	// Wait to merge until we've gone 2 below the node limit, so a node can be efficiently
	// moved around the tree without unnecessary merging and re-subdividing.
	if t.count < t.tqt.nodeLimit-1 && t.branches[0] != nil {
		// This region has shrunk enough that we should merge its children up
		t.leaves = make([]uint64, 0, t.tqt.nodeLimit)
		for i, child := range t.branches {
			t.leaves = append(t.leaves, child.leaves...)
			t.branches[i] = nil
		}
	}

	delete(leafMap, id)
}

func isWithin(loc, topLeft, bottomRight Vector) bool {
	if topLeft.X < bottomRight.X {
		if loc.X < topLeft.X || loc.X >= bottomRight.X {
			return false
		}
	} else {
		if loc.X < topLeft.X && loc.X >= bottomRight.X {
			return false
		}
	}
	if topLeft.Y < bottomRight.Y {
		if loc.Y < topLeft.Y || loc.Y >= bottomRight.Y {
			return false
		}
	} else {
		if loc.Y < topLeft.Y && loc.Y >= bottomRight.Y {
			return false
		}
	}

	return true
}

func regionWithin(rTopLeft, rBottomRight, topLeft, bottomRight Vector) bool {
	if rTopLeft.X < rBottomRight.X {
		if rTopLeft.X >= bottomRight.X || rBottomRight.X <= topLeft.X {
			return false
		}
	} else {
		if rTopLeft.X >= bottomRight.X && rBottomRight.X <= topLeft.X {
			return false
		}
	}
	if rTopLeft.Y < rBottomRight.Y {
		if rTopLeft.Y >= bottomRight.Y || rBottomRight.Y <= topLeft.Y {
			return false
		}
	} else {
		if rTopLeft.Y >= bottomRight.Y && rBottomRight.Y <= topLeft.Y {
			return false
		}
	}

	return true
}

func (t *tree) find(topLeft, bottomRight Vector) []uint64 {
	result := make([]uint64, 0, t.count)

	if t.branches[0] == nil {
		for _, id := range t.leaves {
			if isWithin(t.tqt.leafMap[id], topLeft, bottomRight) {
				result = append(result, id)
			}
		}
	} else {
		for _, region := range t.branches {
			if regionWithin(topLeft, bottomRight, region.topLeft, region.bottomRight) {
				result = append(result, region.find(topLeft, bottomRight)...)
			}
		}
	}

	return result
}

// New creates a new ToroidalQuadtree with the given width and height.
// nodeLimit is the maximum number of elements in a region before subdivision, and must be at least 4.
func New(nodeLimit int, width, height float64) (*ToroidalQuadtree, error) {
	if nodeLimit < 4 {
		return nil, fmt.Errorf("invalid node limit: must be at least 4")
	}

	result := ToroidalQuadtree{
		nodeLimit: nodeLimit,
		leafMap:   make(map[uint64]Vector),
	}
	result.root = &tree{
		tqt:         &result,
		topLeft:     Vector{X: 0.0, Y: 0.0},
		bottomRight: Vector{X: width, Y: height},
		leaves:      make([]uint64, 0, nodeLimit),
	}

	return &result, nil
}

// Width returns the width of the quadtree.
func (t *ToroidalQuadtree) Width() float64 {
	return t.root.bottomRight.X
}

// Height returns the width of the quadtree.
func (t *ToroidalQuadtree) Height() float64 {
	return t.root.bottomRight.Y
}

// Put sets the location of the element with the given ID.
// If the element is already within the quadtree, moves the element from its old location.
func (t *ToroidalQuadtree) Put(id uint64, p Vector) bool {
	if p.X < 0.0 || p.Y < 0.0 || p.X >= t.root.bottomRight.X || p.Y >= t.root.bottomRight.Y {
		return false
	}

	// If this ID is already in the tree, remove it.
	if _, ok := t.leafMap[id]; ok {
		t.Remove(id)
	}

	// Add the leaf to the leaf map and into the tree.
	t.leafMap[id] = p
	t.root.put(id, p)
	return true
}

// Remove takes the element with the given ID out of the quadtree.
func (t *ToroidalQuadtree) Remove(id uint64) bool {
	// If this ID is already not in the tree, return early.
	if _, ok := t.leafMap[id]; !ok {
		return false
	}
	t.root.remove(id, t.leafMap)
	delete(t.leafMap, id)
	return true
}

// Get returns the location of the element with the given ID.
func (t *ToroidalQuadtree) Get(id uint64) *Vector {
	if loc, ok := t.leafMap[id]; ok {
		return &loc
	}
	return nil
}

// Find returns all the elements by ID within the given region.
// NOTE: A region's location interval is closed on the left side and open on the right.
// The region ((0.0, 0.0), (1.0, 1.0)) contains the point (0.0, 0.0), but not (1.0, 1.0).
// Both corners of the selected region must be within the bounds of the quadtree.
func (t *ToroidalQuadtree) Find(topLeft, bottomRight Vector) []uint64 {
	if topLeft.X < 0.0 || topLeft.Y < 0.0 || topLeft.X > t.root.bottomRight.X || topLeft.Y > t.root.bottomRight.Y {
		return nil
	}
	if bottomRight.X < 0.0 || bottomRight.Y < 0.0 || bottomRight.X > t.root.bottomRight.X || bottomRight.Y > t.root.bottomRight.Y {
		return nil
	}

	return t.root.find(topLeft, bottomRight)
}

// DebugString prints out the tree representation.
func (t *ToroidalQuadtree) DebugString() string {
	var sb strings.Builder
	type node struct {
		level int
		tree  *tree
	}
	var q []node
	q = append(q, node{0, t.root})
	for len(q) > 0 {
		next := q[0]
		q = q[1:]
		if next.tree.branches[0] != nil {
			for _, subtree := range next.tree.branches {
				q = append(q, node{next.level + 1, subtree})
			}
		}

		for i := 0; i < next.level-2; i++ {
			sb.WriteString("  ")
		}
		if next.level > 0 {
			sb.WriteString("+-")
		}
		sb.WriteString(fmt.Sprintf("(%v), (%v)\n", next.tree.topLeft, next.tree.bottomRight))

		for _, leaf := range next.tree.leaves {
			for i := 0; i < next.level; i++ {
				sb.WriteString("  ")
			}
			sb.WriteString(fmt.Sprintf("+-%v: (%v)\n", leaf, t.leafMap[leaf]))
		}
	}
	return sb.String()
}
