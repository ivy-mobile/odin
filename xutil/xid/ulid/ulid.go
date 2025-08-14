package ulid

import (
	"crypto/rand"
	"sync"

	"github.com/oklog/ulid/v2"
)

var (
	entropy *ulid.MonotonicEntropy
	mux     sync.Mutex // ulid 本身是非并发安全的
)

func init() {
	entropy = ulid.Monotonic(rand.Reader, 0) // 密码学安全的随机数生成器
}

func NextID() (string, error) {

	mux.Lock()
	defer mux.Unlock()

	id, err := ulid.New(ulid.Now(), entropy)
	if err != nil {
		return "", err
	}
	return id.String(), err
}
