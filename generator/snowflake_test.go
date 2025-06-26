package generator_test

import (
	"testing"

	"github.com/ivy-mobile/odin/generator"
)

func TestSnowflake(t *testing.T) {

	m := make(map[string]struct{})
	g1 := generator.NewSnowflakeIDGenerator(1)
	for range 10000 {
		m[g1.GenRoomID()] = struct{}{}
	}
	g2 := generator.NewSnowflakeIDGenerator(2)
	for range 10000 {
		id := g2.GenRoomID()
		if _, ok := m[id]; ok {
			t.Errorf("duplicate1 id: %s", id)
		}
	}
}

func BenchmarkSnowflakeIDGenerator_GenerateID(b *testing.B) {
	g := generator.NewSnowflakeIDGenerator(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GenRoomID()
	}
}
