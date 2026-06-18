package nacos

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"net"
	"net/url"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/ivy-mobile/odin/registry"
)

var ErrServiceInstanceNameEmpty = errors.New("meta/nacos: ServiceInstance.Name can not be empty")

const (
	fieldKind    = "kind"
	fieldVersion = "version"
	fieldID      = "id"
	fieldWeight  = "weight"
)

// Registry 是 nacos 注册中心实现
type Registry struct {
	opts options
	cli  naming_client.INamingClient
}

var _ registry.Registry = (*Registry)(nil)

// New 创建 nacos 注册中心
func New(cli naming_client.INamingClient, opts ...Option) (r *Registry) {
	op := defaultOptions()
	for _, option := range opts {
		option(&op)
	}
	return &Registry{
		opts: op,
		cli:  cli,
	}
}

func (r *Registry) ID() string {
	return "nacos"
}

func toServiceInstance(in model.Instance, defaultKind string, defaultWeight float64) *registry.ServiceInstance {
	metadata := maps.Clone(in.Metadata)
	if metadata == nil {
		metadata = make(map[string]string, 1)
	}
	kind := defaultKind
	weight := defaultWeight
	if k, ok := metadata[fieldKind]; ok {
		kind = k
	}
	if in.Weight > 0 {
		weight = in.Weight
	}

	r := &registry.ServiceInstance{
		ID:        in.InstanceId,
		Name:      in.ServiceName,
		Version:   metadata[fieldVersion],
		Metadata:  metadata,
		Endpoints: []string{kind + "://" + net.JoinHostPort(in.Ip, strconv.Itoa(int(in.Port)))},
	}
	r.Metadata[fieldWeight] = strconv.FormatInt(int64(math.Ceil(weight)), 10)
	return r
}

// Register 注册服务
func (r *Registry) Register(_ context.Context, si *registry.ServiceInstance) error {
	if si.Name == "" {
		return ErrServiceInstanceNameEmpty
	}
	for _, endpoint := range si.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		weight := r.opts.weight
		var rmd map[string]string
		if si.Metadata == nil {
			rmd = map[string]string{
				fieldKind:    u.Scheme,
				fieldVersion: si.Version,
			}
		} else {
			rmd = maps.Clone(si.Metadata)
			rmd[fieldKind] = u.Scheme
			rmd[fieldVersion] = si.Version
			rmd[fieldID] = si.ID
			if w, ok := si.Metadata[fieldWeight]; ok {
				weight, err = strconv.ParseFloat(w, 64)
				if err != nil {
					weight = r.opts.weight
				}
			}
		}
		success, e := r.cli.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: si.Name,
			Weight:      weight,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    rmd,
			ClusterName: r.opts.cluster,
			GroupName:   r.opts.group,
		})
		if e != nil {
			return fmt.Errorf("RegisterInstance err %v,%v", e, endpoint)
		}
		if !success {
			return fmt.Errorf("RegisterInstance failed,%v", endpoint)
		}
	}
	return nil
}

// Deregister 注销服务
func (r *Registry) Deregister(_ context.Context, service *registry.ServiceInstance) error {
	for _, endpoint := range service.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		success, err := r.cli.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: service.Name,
			GroupName:   r.opts.group,
			Cluster:     r.opts.cluster,
			Ephemeral:   true,
		})
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("DeregisterInstance failed,%v", endpoint)
		}
	}
	return nil
}

// Watch 按服务名创建 watcher
func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newWatcher(ctx, r.cli, serviceName, r.opts.group, r.opts.kind, r.opts.weight, []string{r.opts.cluster})
}

// GetService 按服务名获取实例列表
func (r *Registry) GetService(_ context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	res, err := r.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   r.opts.group,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(res))
	for _, in := range res {
		items = append(items, toServiceInstance(in, r.opts.kind, r.opts.weight))
	}
	return items, nil
}
