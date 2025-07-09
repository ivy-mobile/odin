package etcd

import (
	"fmt"

	"github.com/ivy-mobile/odin/enum"
	"github.com/ivy-mobile/odin/registry"
)

// 构建实例ID
func makeInsID(ins *registry.ServiceInstance) string {
	if ins.Kind == enum.NodeType_Game {
		return fmt.Sprintf("%s-%s-%s-%s", ins.Kind, ins.Name, ins.Alias, ins.ID)
	}
	return fmt.Sprintf("%s-%s-%s", ins.Kind, ins.Name, ins.ID)
}
