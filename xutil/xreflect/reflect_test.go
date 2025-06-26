package xreflect_test

import (
	"testing"

	"github.com/ivy-mobile/odin/xutil/xreflect"
)

func TestIsNil(t *testing.T) {
	var b1 bool
	var b2 *bool
	t.Log(xreflect.IsNil(b1))
	t.Log(xreflect.IsNil(&b1))
	t.Log(xreflect.IsNil(b2))
	t.Log(xreflect.IsNil(&b2))
}
