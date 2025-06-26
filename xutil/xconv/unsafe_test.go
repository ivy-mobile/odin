package xconv_test

import (
	"fmt"
	"testing"

	"github.com/ivy-mobile/odin/xutil/xconv"
)

func TestBytesToString(t *testing.T) {
	b := []byte("abc")

	s := xconv.BytesToString(b)
	fmt.Printf("%s\n", s)
	fmt.Printf("%p\n", &b)
	fmt.Printf("%p\n", &s)
}

func BenchmarkBytesToString(b *testing.B) {

	for i := 0; i < b.N; i++ {
		xconv.BytesToString([]byte("abc"))
	}
}
