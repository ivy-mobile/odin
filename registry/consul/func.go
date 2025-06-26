package consul

import (
	"fmt"

	"github.com/ivy-mobile/odin/registry"
)

// 构建实例ID
func makeId(ins *registry.ServiceInstance) string {
	return fmt.Sprintf("%s-%s", ins.Kind, ins.ID)
}
