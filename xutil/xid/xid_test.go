package xid

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestSonyflake(t *testing.T) {

	total := 1000
	g, ctx := errgroup.WithContext(context.Background())

	mux := sync.Mutex{}
	idMap := make(map[int64]struct{})
	for range total {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				id, err := Sonyflake()
				if err != nil {
					return err
				}

				mux.Lock()
				idMap[id] = struct{}{}
				mux.Unlock()

				return nil
			}
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, total, len(idMap))
}

func BenchmarkSonyflake(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Sonyflake()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestSnowflake(t *testing.T) {

	total := 1000
	g, ctx := errgroup.WithContext(context.Background())

	mux := sync.Mutex{}
	idMap := make(map[string]struct{})
	for range total {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				id := Snowflake()
				mux.Lock()
				idMap[id] = struct{}{}
				mux.Unlock()

				return nil
			}
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, total, len(idMap))
}

func BenchmarkSnowflake(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Snowflake()
	}
}

func TestUlid(t *testing.T) {

	total := 1000
	g, ctx := errgroup.WithContext(context.Background())

	mux := sync.Mutex{}
	idMap := make(map[string]struct{})
	for range total {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				id, err := ULID()
				if err != nil {
					return err
				}
				mux.Lock()
				idMap[id] = struct{}{}
				mux.Unlock()

				return nil
			}
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, total, len(idMap))
}

func BenchmarkUlid(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ULID()
		if err != nil {
			b.Fatal(err)
		}
	}
}
