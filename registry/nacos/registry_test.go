package nacos_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/registry"
	"github.com/ivy-mobile/odin/registry/nacos"
	"github.com/ivy-mobile/odin/xutil/xconv"
	"github.com/ivy-mobile/odin/xutil/xnet"

	"golang.org/x/sync/errgroup"
)

const (
	port        = 3553
	serviceName = "gate"
)

var reg = nacos.NewRegistry(
	nacos.WithUrls("http://43.153.4.107:18848/nacos"),
	nacos.WithGroupName("party-pop-games"),
	nacos.WithNamespaceId("zhaobin"),
)

func TestRegistry_Register1(t *testing.T) {
	host, err := xnet.ExternalIP()
	if err != nil {
		t.Fatal(err)
	}

	cnt := 0
	ctx := context.Background()
	ins := &registry.ServiceInstance{
		ID:       "test-1",
		Name:     serviceName,
		Kind:     "NODE",
		Alias:    "login-server",
		State:    "WORK",
		Endpoint: fmt.Sprintf("grpc://%s:%d", host, port),
	}

	for {
		if cnt%2 == 0 {
			ins.State = "WORK"
		} else {
			ins.State = "BUSY"
		}

		if err = reg.Register(ctx, ins); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("register: %v", ins)
		}

		cnt++

		time.Sleep(2 * time.Second)
	}
}

func TestRegistry_Register2(t *testing.T) {
	host, err := xnet.ExternalIP()
	if err != nil {
		t.Fatal(err)
	}

	if err = reg.Register(context.Background(), &registry.ServiceInstance{
		ID:       "test-2",
		Name:     serviceName,
		Kind:     "NODE",
		State:    "WORK",
		Endpoint: fmt.Sprintf("grpc://%s:%d", host, port),
	}); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
	}()

	time.Sleep(30 * time.Second)
}

func TestRegistry_Services(t *testing.T) {
	services, err := reg.Services(context.Background(), serviceName)
	if err != nil {
		t.Fatal(err)
	}

	for _, service := range services {
		t.Logf("%+v", service)
	}
}

func TestRegistry_Watch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	watcher1, err := reg.Watch(ctx, serviceName)
	cancel()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	watcher2, err := reg.Watch(ctx, serviceName)
	cancel()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			services, err := watcher1.Next()
			if err != nil {
				t.Errorf("goroutine 1: %v", err)
				return
			}

			for _, service := range services {
				t.Logf("goroutine 1: %+v", service)
			}
		}
	}()

	go func() {
		for {
			services, err := watcher2.Next()
			if err != nil {
				t.Errorf("goroutine 2: %v", err)
				return
			}

			for _, service := range services {
				t.Logf("goroutine 2: %+v", service)
			}
		}
	}()

	select {}
}

func TestMultipleNodeRegister(t *testing.T) {
	for i := range 5 {
		go func(i int) {
			n := newNode(xconv.String(i))
			n.start()
		}(i)
	}

	time.Sleep(10 * time.Second)
}

const (
	defaultTimeout = 3 * time.Second // 默认超时时间
)

type node struct {
	id        string
	ctx       context.Context
	registry  registry.Registry
	instances []*registry.ServiceInstance
}

func newNode(id string) *node {
	n := &node{}
	n.id = id
	n.ctx = context.Background()
	n.registry = nacos.NewRegistry()
	n.instances = make([]*registry.ServiceInstance, 0)

	n.instances = append(n.instances, &registry.ServiceInstance{
		ID:    id,
		Name:  "NODE",
		Kind:  "NODE",
		Alias: fmt.Sprintf("node-%s", id),
		State: "WORK",
		Routes: []registry.Route{
			{ID: 1, Stateful: true, Internal: false},
			{ID: 2, Stateful: true, Internal: false},
			{ID: 3, Stateful: true, Internal: false},
			{ID: 4, Stateful: true, Internal: false},
		},
		Endpoint: fmt.Sprintf("grpc://%s:%d", id, port),
	})

	return n
}

func (n *node) start() {
	n.watch()

	if err := n.register(); err != nil {
		log.Fatalf("register cluster instances failed: %v", err)
	}

}

// 执行注册操作
func (n *node) register() error {
	eg, ctx := errgroup.WithContext(n.ctx)

	for i := range n.instances {
		instance := n.instances[i]
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
			return n.registry.Register(ctx, instance)
		})
	}

	return eg.Wait()
}

func (n *node) watch() {
	ctx, cancel := context.WithTimeout(n.ctx, 3*time.Second)
	watcher, err := n.registry.Watch(ctx, "NODE")
	cancel()
	if err != nil {
		log.Fatalf("the dispatcher instance watch failed: %v", err)
	}

	go func() {
		defer watcher.Stop()
		for {
			select {
			case <-n.ctx.Done():
				return
			default:
				// exec watch
			}

			services, err := watcher.Next()
			if err != nil {
				continue
			}

			fmt.Printf("node: %v services: %v\n", n.id, len(services))

			for _, service := range services {
				fmt.Printf("service id: %v\n", service.ID)
			}

			fmt.Println()
		}
	}()
}
