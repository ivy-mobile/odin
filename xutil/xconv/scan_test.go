package xconv_test

import (
	"testing"

	"github.com/ivy-mobile/odin/xutil/xconv"
)

func TestScanStruct(t *testing.T) {

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := `{"name":"Charlie", "age":1}`
	var user User
	err := xconv.Scan(data, &user)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
}

func TestScanStructSlice(t *testing.T) {

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := `[{"name":"Charlie1", "age":1},{"name":"Charlie2", "age":2}]`
	var users []User
	err := xconv.Scan(data, &users)
	if err != nil {
		t.Error(err)
	}
	t.Log(users)
}
