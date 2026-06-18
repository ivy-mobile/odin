package nacos

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ivy-mobile/odin/registry"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type mockNamingClient struct {
	registerParams   []vo.RegisterInstanceParam
	registerErr      error
	registerFailed   bool
	deregisterParams []vo.DeregisterInstanceParam
	deregisterErr    error
	deregisterFailed bool
	selectParam      vo.SelectInstancesParam
	selectInstances  []model.Instance
	selectErr        error
	getServiceParam  vo.GetServiceParam
	service          model.Service
	getServiceErr    error
	subscribeParam   *vo.SubscribeParam
	subscribeErr     error
	unsubscribeParam *vo.SubscribeParam
	unsubscribeErr   error
}

func (m *mockNamingClient) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	m.registerParams = append(m.registerParams, param)
	if m.registerFailed {
		return false, m.registerErr
	}
	return m.registerErr == nil, m.registerErr
}

func (m *mockNamingClient) BatchRegisterInstance(vo.BatchRegisterInstanceParam) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockNamingClient) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	m.deregisterParams = append(m.deregisterParams, param)
	if m.deregisterFailed {
		return false, m.deregisterErr
	}
	return m.deregisterErr == nil, m.deregisterErr
}

func (m *mockNamingClient) UpdateInstance(vo.UpdateInstanceParam) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockNamingClient) GetService(param vo.GetServiceParam) (model.Service, error) {
	m.getServiceParam = param
	return m.service, m.getServiceErr
}

func (m *mockNamingClient) SelectAllInstances(vo.SelectAllInstancesParam) ([]model.Instance, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNamingClient) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	m.selectParam = param
	return m.selectInstances, m.selectErr
}

func (m *mockNamingClient) SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNamingClient) Subscribe(param *vo.SubscribeParam) error {
	m.subscribeParam = param
	return m.subscribeErr
}

func (m *mockNamingClient) Unsubscribe(param *vo.SubscribeParam) error {
	m.unsubscribeParam = param
	return m.unsubscribeErr
}

func (m *mockNamingClient) GetAllServicesInfo(vo.GetAllServiceInfoParam) (model.ServiceList, error) {
	return model.ServiceList{}, errors.New("not implemented")
}

func (m *mockNamingClient) ServerHealthy() bool {
	return true
}

func (m *mockNamingClient) CloseClient() {}

func TestRegistry_RegisterBuildsNacosParams(t *testing.T) {
	client := &mockNamingClient{}
	r := New(client, Group("CUSTOM_GROUP"), Cluster("blue"), Weight(200))
	metadata := map[string]string{"idc": "shanghai", "weight": "12.3"}

	err := r.Register(context.Background(), &registry.ServiceInstance{
		ID:        "node-1",
		Name:      "game.grpc",
		Version:   "v1.0.0",
		Metadata:  metadata,
		Endpoints: []string{"grpc://127.0.0.1:9000", "http://127.0.0.2:8000"},
	})
	if err != nil {
		t.Fatalf("Register error = %v", err)
	}
	if len(client.registerParams) != 2 {
		t.Fatalf("RegisterInstance calls = %d, want 2", len(client.registerParams))
	}

	wantFirst := vo.RegisterInstanceParam{
		Ip:          "127.0.0.1",
		Port:        9000,
		ServiceName: "game.grpc",
		Weight:      12.3,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		ClusterName: "blue",
		GroupName:   "CUSTOM_GROUP",
		Metadata: map[string]string{
			"id":      "node-1",
			"idc":     "shanghai",
			"kind":    "grpc",
			"version": "v1.0.0",
			"weight":  "12.3",
		},
	}
	if !reflect.DeepEqual(client.registerParams[0], wantFirst) {
		t.Fatalf("first RegisterInstance param = %#v, want %#v", client.registerParams[0], wantFirst)
	}
	if got := client.registerParams[1]; got.ServiceName != "game.grpc" || got.Metadata["kind"] != "http" {
		t.Fatalf("second RegisterInstance param = %#v", got)
	}
	if !reflect.DeepEqual(metadata, map[string]string{"idc": "shanghai", "weight": "12.3"}) {
		t.Fatalf("Register mutated input metadata: %#v", metadata)
	}
}

func TestRegistry_RegisterUsesServiceNameAsNacosServiceName(t *testing.T) {
	client := &mockNamingClient{}
	r := New(client)

	err := r.Register(context.Background(), &registry.ServiceInstance{
		Name:      "game",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://127.0.0.1:9000"},
	})
	if err != nil {
		t.Fatalf("Register error = %v", err)
	}
	if got := client.registerParams[0].ServiceName; got != "game" {
		t.Fatalf("ServiceName = %q, want %q", got, "game")
	}
}

func TestRegistry_RegisterReturnsErrorWhenNacosReturnsFalse(t *testing.T) {
	client := &mockNamingClient{registerFailed: true}
	r := New(client)

	err := r.Register(context.Background(), &registry.ServiceInstance{
		Name:      "game",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://127.0.0.1:9000"},
	})
	if err == nil {
		t.Fatal("Register error = nil, want error")
	}
}

func TestRegistry_DeregisterBuildsNacosParams(t *testing.T) {
	client := &mockNamingClient{}
	r := New(client, Group("CUSTOM_GROUP"), Cluster("blue"))

	err := r.Deregister(context.Background(), &registry.ServiceInstance{
		Name:      "game.grpc",
		Endpoints: []string{"grpc://127.0.0.1:9000"},
	})
	if err != nil {
		t.Fatalf("Deregister error = %v", err)
	}
	want := vo.DeregisterInstanceParam{
		Ip:          "127.0.0.1",
		Port:        9000,
		ServiceName: "game.grpc",
		GroupName:   "CUSTOM_GROUP",
		Cluster:     "blue",
		Ephemeral:   true,
	}
	if !reflect.DeepEqual(client.deregisterParams[0], want) {
		t.Fatalf("DeregisterInstance param = %#v, want %#v", client.deregisterParams[0], want)
	}
}

func TestRegistry_DeregisterReturnsErrorWhenNacosReturnsFalse(t *testing.T) {
	client := &mockNamingClient{deregisterFailed: true}
	r := New(client)

	err := r.Deregister(context.Background(), &registry.ServiceInstance{
		Name:      "game",
		Endpoints: []string{"grpc://127.0.0.1:9000"},
	})
	if err == nil {
		t.Fatal("Deregister error = nil, want error")
	}
}

func TestRegistry_GetServiceMapsInstances(t *testing.T) {
	metadata := map[string]string{"version": "v1.0.0", "kind": "http"}
	client := &mockNamingClient{
		selectInstances: []model.Instance{
			{
				InstanceId:  "i-1",
				Ip:          "127.0.0.1",
				Port:        8000,
				Weight:      12.2,
				ServiceName: "CUSTOM_GROUP@@game.http",
				Metadata:    metadata,
			},
			{
				InstanceId:  "i-2",
				Ip:          "127.0.0.2",
				Port:        9000,
				ServiceName: "CUSTOM_GROUP@@game.tcp",
			},
		},
	}
	r := New(client, Group("CUSTOM_GROUP"), Kind("tcp"), Weight(33))

	got, err := r.GetService(context.Background(), "game.http")
	if err != nil {
		t.Fatalf("GetService error = %v", err)
	}
	if client.selectParam.ServiceName != "game.http" || client.selectParam.GroupName != "CUSTOM_GROUP" || !client.selectParam.HealthyOnly {
		t.Fatalf("SelectInstances param = %#v", client.selectParam)
	}

	want := []*registry.ServiceInstance{
		{
			ID:        "i-1",
			Name:      "CUSTOM_GROUP@@game.http",
			Version:   "v1.0.0",
			Metadata:  map[string]string{"version": "v1.0.0", "kind": "http", "weight": "13"},
			Endpoints: []string{"http://127.0.0.1:8000"},
		},
		{
			ID:        "i-2",
			Name:      "CUSTOM_GROUP@@game.tcp",
			Metadata:  map[string]string{"weight": "33"},
			Endpoints: []string{"tcp://127.0.0.2:9000"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetService got = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(metadata, map[string]string{"version": "v1.0.0", "kind": "http"}) {
		t.Fatalf("GetService mutated source metadata: %#v", metadata)
	}
}

func TestRegistry_WatchMapsServiceAndUnsubscribes(t *testing.T) {
	metadata := map[string]string{"version": "v1.0.0", "kind": "grpc"}
	client := &mockNamingClient{
		service: model.Service{
			Hosts: []model.Instance{{
				InstanceId:  "i-1",
				Ip:          "127.0.0.1",
				Port:        9000,
				Weight:      9.1,
				ServiceName: "CUSTOM_GROUP@@game.grpc",
				Metadata:    metadata,
			}},
		},
	}
	r := New(client, Group("CUSTOM_GROUP"), Cluster("blue"), Weight(44))

	watch, err := r.Watch(context.Background(), "game.grpc")
	if err != nil {
		t.Fatalf("Watch error = %v", err)
	}
	if client.subscribeParam.ServiceName != "game.grpc" ||
		client.subscribeParam.GroupName != "CUSTOM_GROUP" ||
		!reflect.DeepEqual(client.subscribeParam.Clusters, []string{"blue"}) {
		t.Fatalf("Subscribe param = %#v", client.subscribeParam)
	}

	got, err := watch.Next()
	if err != nil {
		t.Fatalf("Next error = %v", err)
	}
	want := []*registry.ServiceInstance{{
		ID:        "i-1",
		Name:      "CUSTOM_GROUP@@game.grpc",
		Version:   "v1.0.0",
		Metadata:  map[string]string{"version": "v1.0.0", "kind": "grpc", "weight": "10"},
		Endpoints: []string{"grpc://127.0.0.1:9000"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Next got = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(metadata, map[string]string{"version": "v1.0.0", "kind": "grpc"}) {
		t.Fatalf("Next mutated source metadata: %#v", metadata)
	}

	if err := watch.Stop(); err != nil {
		t.Fatalf("Stop error = %v", err)
	}
	if client.unsubscribeParam != client.subscribeParam {
		t.Fatalf("Unsubscribe param = %#v, want subscribe param %#v", client.unsubscribeParam, client.subscribeParam)
	}
}
