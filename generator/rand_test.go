package generator_test

import (
	"testing"

	"github.com/ivy-mobile/odin/generator"
)

func TestRandGenerator(t *testing.T) {
	g := generator.NewRandGenerator()
	m := make(map[string]struct{})
	for i := 0; i < 10000; i++ {
		id := g.GenRoomID()
		if _, ok := m[id]; ok {
			t.Errorf("duplicate id: %s", id)
		}
		m[id] = struct{}{}
		// t.Logf(g.GenerateID())
	}
}
