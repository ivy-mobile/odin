package nacos

import (
	"fmt"

	"github.com/ivy-mobile/odin/registry"
)

// 构建实例ID
func makeInsID(ins *registry.ServiceInstance) string {
	return fmt.Sprintf("%s-%s", ins.Kind, ins.ID)
}
