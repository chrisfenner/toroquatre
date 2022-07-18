% toroquatre
toroquatre is a simple toroidal quadtree data structure library.

A quadtree is used for efficient collision detection between lots of objects.
Instead of checking every object's position against every other object's
position, the objects' two-dimensional spatial relationships to each other are
stored in a tree, where each node either contains a limited number of leaf
nodes, or four child nodes dividing the total space of the parent node.

A toroidal space loops around across both the top and bottom edges, like
Pac-Man. If you travel to the right off the right edge of the space, you come
back in on the left side. This library is intended to provide an efficient
quadtree that comprehends this wrapping so callers don't have to.

Objects are represented by `uint64` ids - store whatever metadata you want
somewhere else. The tree understands only positions by ID.

## Usage

Create the quadtree over your space by using `toroquatre.New`:

```go
tree, err := toroquatre.New(nodeLimit, width, height)
```

`nodeLimit` must be at least 4 and may be tuned to your use case.

Put or move objects within the tree using `Put`:

```go
tree.Put(1, toroquatre.Point{X: x, Y: y})
```

Remove them with `Remove`:
```go
tree.Remove(1)
```

Query a rectangular region of the tree with `Find`:
```go
tree.Find(toroquatre.Point{X: x1, Y: y1}, toroquatre.Point{X: x2, Y: y2})
```

It returns all the object IDs within the rectangle bounded by
`(x1, y1, x2, y2)`. `(x1, y1)` is always the top left of the rectangle, and
`(x2, y2)` is always the bottom right. If `(x1, y1)` is to the right of or
below `(x2, y2)`, the rectangle wraps around the sides or top/bottom of the
whole space.