package engine

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/ivy-mobile/odin/conf"
)

// NewNacosNamingClient 新建Nacos naming客户端
func NewNacosNamingClient(cfg conf.NacosConfig) (naming_client.INamingClient, error) {
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId(cfg.Namespace),
		constant.WithTimeoutMs(cfg.RequestTimeout),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(cfg.LogDir),
		constant.WithCacheDir(cfg.CacheDir),
		constant.WithLogLevel(cfg.LogLevel),
		constant.WithBeatInterval(5000),
	)
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      cfg.IpAddr,
			Port:        cfg.Port,
			ContextPath: cfg.Path,
		},
	}

	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("new nacos naming client: %w", err)
	}
	return client, nil
}
