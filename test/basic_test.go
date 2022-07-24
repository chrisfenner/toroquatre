package tqttest

import (
	"fmt"
	"math"
	"testing"

	"github.com/chrisfenner/toroquatre"
)

func UnorderedEqual(list1, list2 []uint64) bool {
	if len(list1) != len(list2) {
		return false
	}
	set1 := make(map[uint64]bool)
	for _, e := range list1 {
		set1[e] = true
	}
	set2 := make(map[uint64]bool)
	for _, e := range list2 {
		set2[e] = true
	}
	for e, _ := range set1 {
		if _, ok := set2[e]; !ok {
			return false
		}
	}
	return true
}

func TestAddSomeElements(t *testing.T) {
	for _, nodeLimit := range []int{4, 5, 6, 7} {
		t.Run(fmt.Sprintf("nodeLimit=%d", nodeLimit), func(t *testing.T) {
			tree, err := toroquatre.New(nodeLimit, 6.0, 10.0)
			if err != nil {
				t.Fatalf("error from New: %v", err)
			}
			points := []toroquatre.Vector{
				{X: 1.0, Y: 0.0},
				{X: 2.0, Y: 1.0},
				{X: 4.0, Y: 1.0},
				{X: 5.0, Y: 1.0},
				{X: 3.0, Y: 3.0},
				{X: 5.0, Y: 6.0},
				{X: 1.0, Y: 3.0},
				{X: 5.0, Y: 9.0},
				{X: 0.0, Y: 4.0},
			}
			for id, loc := range points {
				tree.Put(uint64(id), loc)
			}
			for id, loc := range points {
				if got := tree.Get(uint64(id)); got == nil || *got != loc {
					t.Errorf("point with id %v had location %v, want %v", id, got, &loc)
				}
			}
			if got := tree.Get(1000); got != nil {
				t.Errorf("un-added point had location %v, want nil", got)
			}
			// Look for each element in a very small area near itself
			for id, loc := range points {
				topLeft := loc
				topLeft.X -= 0.01
				if topLeft.X < 0.0 {
					topLeft.X += tree.Width()
				}
				topLeft.Y -= 0.01
				if topLeft.Y < 0.0 {
					topLeft.Y += tree.Height()
				}
				bottomRight := loc
				bottomRight.X += 0.01
				if bottomRight.X >= tree.Width() {
					bottomRight.X -= tree.Width()
				}
				bottomRight.Y += 0.01
				if bottomRight.Y >= tree.Height() {
					bottomRight.Y -= tree.Height()
				}
				results := tree.Find(topLeft, bottomRight)
				if len(results) != 1 {
					t.Errorf("wanted to find 1 point in (%v, %v), got %v", topLeft, bottomRight, len(results))
				}
				if len(results) == 1 && results[0] != uint64(id) {
					t.Errorf("wanted to find %v in (%v, %v), got %v", id, topLeft, bottomRight, results[0])
				}
			}
			// Look for particular elements in specific subregions
			type subregionQuery struct {
				wantIDs     []uint64
				topLeft     toroquatre.Vector
				bottomRight toroquatre.Vector
			}
			for i, q := range []subregionQuery{
				{
					[]uint64{0, 1, 6},
					toroquatre.Vector{X: 0.0, Y: 0.0},
					toroquatre.Vector{X: 2.1, Y: 3.1},
				},
				{
					[]uint64{3, 5, 6, 8},
					toroquatre.Vector{X: 4.5, Y: 0.5},
					toroquatre.Vector{X: 1.5, Y: 6.5},
				},
				{
					[]uint64{0, 3, 7},
					toroquatre.Vector{X: 4.5, Y: 8.5},
					toroquatre.Vector{X: 1.5, Y: 1.5},
				},
				{
					[]uint64{2, 3, 5, 7},
					toroquatre.Vector{X: 3.5, Y: 5.5},
					toroquatre.Vector{X: 5.5, Y: 1.5},
				},
			} {
				t.Run(fmt.Sprintf("subregions_%v", i), func(t *testing.T) {
					got := tree.Find(q.topLeft, q.bottomRight)
					if !UnorderedEqual(q.wantIDs, got) {
						t.Errorf("within %v to %v, want %v, got %v", q.topLeft, q.bottomRight, q.wantIDs, got)
					}
				})
			}
		})
	}
}

func TestLotsOfElements(t *testing.T) {
	for _, nodeLimit := range []int{4, 5, 6, 7} {
		t.Run(fmt.Sprintf("nodeLimit=%d", nodeLimit), func(t *testing.T) {
			tree, err := toroquatre.New(nodeLimit, 10.0, 10.0)
			if err != nil {
				t.Fatalf("error from New: %v", err)
			}
			for i := 0; i < 20000; i++ {
				angle := float64(i) * math.Pi / 10000
				// Put the elemenet into one spot and then move it.
				tree.Put(uint64(i), toroquatre.Vector{X: 5.0 + (3.0 * math.Cos(angle)), Y: 5.0 + (3.0 * math.Sin(angle))})
				tree.Put(uint64(i), toroquatre.Vector{X: 5.0 + (4.0 * math.Cos(angle)), Y: 5.0 + (4.0 * math.Sin(angle))})
			}
			one := tree.Find(toroquatre.Vector{X: 0.9999, Y: 4.9999}, toroquatre.Vector{X: 1.0001, Y: 5.0001})
			if len(one) != 1 {
				t.Errorf("wanted 1 elements got %v", one)
			}
			quarter := tree.Find(toroquatre.Vector{X: 0.0, Y: 0.0}, toroquatre.Vector{X: 5.0, Y: 5.0})
			if len(quarter) != 5000 {
				t.Errorf("wanted 5000 elements got %v", len(quarter))
			}
		})
	}
}

func TestAddRemove(t *testing.T) {
	for _, nodeLimit := range []int{4, 5, 6, 7} {
		t.Run(fmt.Sprintf("nodeLimit=%d", nodeLimit), func(t *testing.T) {
			tree, err := toroquatre.New(nodeLimit, 10.0, 10.0)
			if err != nil {
				t.Fatalf("error from New: %v", err)
			}
			tree.Put(0, toroquatre.Vector{X: 1.0, Y: 2.0})
			if !tree.Remove(0) {
				t.Errorf("want true got false")
			}
			if tree.Remove(0) {
				t.Errorf("want false got true")
			}
			tree.Put(0, toroquatre.Vector{X: 1.0, Y: 2.0})
			tree.Put(0, toroquatre.Vector{X: 1.0, Y: 3.0})
			if found := tree.Find(toroquatre.Vector{X: 0.99, Y: 1.99}, toroquatre.Vector{X: 1.01, Y: 2.01}); len(found) != 0 {
				t.Errorf("want nil got %v", found)
			}
			if found := tree.Find(toroquatre.Vector{X: 0.99, Y: 2.99}, toroquatre.Vector{X: 1.01, Y: 3.01}); len(found) != 1 {
				t.Errorf("want one item got %v", found)
			}
			if !tree.Remove(0) {
				t.Errorf("want true got false")
			}
			if tree.Remove(0) {
				t.Errorf("want false got true")
			}
		})
	}
}
