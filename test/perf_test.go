package tqttest

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/chrisfenner/toroquatre"
)

func BenchmarkFind(b *testing.B) {
	for _, nodeLimit := range []int{4, 8, 16, 32, 64} {
		b.Run(fmt.Sprintf("nodeLimit=%d", nodeLimit), func(b *testing.B) {
			tree, err := toroquatre.New(nodeLimit, 1.0, 1.0)
			if err != nil {
				b.Fatalf("error from toroquatre.New: %v", err)
			}
			for i := 0; i < 1000000; i++ {
				tree.Put(uint64(i), toroquatre.Point{X: rand.Float64(), Y: rand.Float64()})
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				loc := tree.Get(uint64(rand.Int63n(1000000)))
				locTopLeft := toroquatre.Point{X: loc.X - 0.01, Y: loc.Y - 0.01}
				if locTopLeft.X < 0.0 {
					locTopLeft.X += 1.0
				}
				if locTopLeft.Y < 0.0 {
					locTopLeft.Y += 1.0
				}
				locBottomRight := toroquatre.Point{X: loc.X + 0.01, Y: loc.Y + 0.01}
				if locBottomRight.X >= 1.0 {
					locBottomRight.X -= 1.0
				}
				if locBottomRight.Y >= 1.0 {
					locBottomRight.Y -= 1.0
				}
				tree.Find(locTopLeft, locBottomRight)
			}
		})
	}
}
