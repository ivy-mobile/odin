package xfile

import "testing"

func TestJoinFilename(t *testing.T) {

	t.Logf("result: %s", JoinFilename("/a/b/c.log", "-", "1"))
	t.Logf("result: %s", JoinFilename("a/b/c.log", "-", "1"))
	t.Logf("result: %s", JoinFilename("/c.log", "-", "1"))
	t.Logf("result: %s", JoinFilename("c.log", "-", "1"))
	t.Logf("result: %s", JoinFilename("c", "-", "1"))
}
