package instance

import (
	"errors"
	"strings"

	consul "github.com/hashicorp/consul/api"
	pb "github.com/opencopilot/core/core"
)

// Instance is a open-copilot managed instance
type Instance struct {
	ID       string
	Provider *Provider
	Services []string
	Owner    string
	Device   string
}

// Provider is an instance provider (such as Packet)
type Provider struct {
	provider pb.Provider
}

func (p Provider) String() (string, error) {
	return p.provider.String(), nil
}

// NewInstance returns a new instance
func NewInstance(consulClient consul.Client, id string) (*Instance, error) {
	i := Instance{
		ID: id,
	}
	instance, err := i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (i *Instance) instancePrefix() string {
	return "instances/" + i.ID
}

func (i *Instance) servicesPrefix() string {
	return "instances/" + i.ID + "/services/"
}

// NewProvider returns a provider
func NewProvider(providerName string) (*Provider, error) {

	p, ok := pb.Provider_value[providerName]
	if !ok {
		return nil, errors.New("Invalid providerName")
	}

	provider := &Provider{
		provider: pb.Provider(p),
	}
	return provider, nil
}

// CreateInstanceRequest describes the params for creating an instance
type CreateInstanceRequest struct {
	ID       string
	Provider string
	Owner    string
	Device   string
}

// ToMessage converts an instance to something that can be sent back over gRPC
func (i *Instance) ToMessage() (*pb.Instance, error) {
	return &pb.Instance{
		Id:       i.ID,
		Owner:    i.Owner,
		Provider: i.Provider.provider,
		Device:   i.Device,
	}, nil
}

// GetInstance gets instance info
func (i *Instance) GetInstance(consulClient consul.Client) (*Instance, error) {
	kv := consulClient.KV()

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVGet,
			Key:  i.instancePrefix() + "/provider",
		},
		&consul.KVTxnOp{
			Verb: consul.KVGet,
			Key:  i.instancePrefix() + "/owner",
		},
		&consul.KVTxnOp{
			Verb: consul.KVGet,
			Key:  i.instancePrefix() + "/device",
		},
	}
	ok, response, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not fetch instance from Consul")
	}

	instanceInfo := make(map[string]string)

	for _, res := range response.Results {
		instanceInfo[res.Key] = string(res.Value[:])
	}

	keys, _, err := kv.Keys(i.servicesPrefix(), "/", nil)
	if err != nil {
		return nil, err
	}

	services := make([]string, 0)

	for _, key := range keys[1:] {
		service := strings.Replace(key, i.servicesPrefix(), "", 1)
		services = append(services, strings.Replace(service, "/", "", 1))
	}

	provider, err := NewProvider(instanceInfo[i.instancePrefix()+"/provider"])
	if err != nil {
		return nil, err
	}

	i.Provider = provider
	i.Owner = instanceInfo[i.instancePrefix()+"/owner"]
	i.Device = instanceInfo[i.instancePrefix()+"/device"]
	i.Services = services

	return i, nil
}

// CreateInstance creates the key/value pairs for a new instance in Consul
func CreateInstance(consulClient consul.Client, instanceParams CreateInstanceRequest) (*Instance, error) {
	kv := consulClient.KV()

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVSet,
			Key:  "instances/" + instanceParams.ID + "/services/",
		},
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + instanceParams.ID + "/provider",
			Value: []byte(instanceParams.Provider),
		},
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + instanceParams.ID + "/owner",
			Value: []byte(instanceParams.Owner),
		},
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + instanceParams.ID + "/device",
			Value: []byte(instanceParams.Device),
		},
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not create instance in Consul")
	}

	i := Instance{ID: instanceParams.ID}
	instance, err := i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func AddService(consulClient consul.Client) (*Instance, error) {
	return nil, nil
}

func RemoveService(consulClient consul.Client) (*Instance, error) {
	return nil, nil
}

// ListInstances lists all instances from Consul
func ListInstances(consulClient consul.Client) ([]*Instance, error) {
	kv := consulClient.KV()
	keys, _, err := kv.Keys("instances/", "/", nil)
	if err != nil {
		return nil, err
	}

	instances := make([]*Instance, 0)

	for _, key := range keys[1:] {
		instance := strings.Replace(key, "instances/", "", 1)
		instanceID := strings.Replace(instance, "/", "", 1)
		i, _ := NewInstance(consulClient, instanceID)
		instances = append(instances, i)
	}

	return instances, nil
}
