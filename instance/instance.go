package instance

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/buger/jsonparser"

	consul "github.com/hashicorp/consul/api"
	"github.com/opencopilot/consulkvjson"
	pb "github.com/opencopilot/core/core"
)

// Instance is a open-copilot managed instance
type Instance struct {
	ID       string
	Provider *Provider
	Services Services
	Owner    string
	Device   string
}

// Service is a managed service
type Service struct {
	Type   string
	Config string
}

// ToMessage serializes a Service for gRPC
func (s *Service) ToMessage() (*pb.Service, error) {
	return &pb.Service{
		Type:   s.Type,
		Config: s.Config,
	}, nil
}

// Services is a list of Service
type Services []*Service

// ToMessage serializes a list of Services for gRPC
func (services Services) ToMessage() ([]*pb.Service, error) {
	s := make([]*pb.Service, 0)
	for _, service := range services {
		serialized, err := service.ToMessage()
		if err != nil {
			return nil, err
		}
		s = append(s, serialized)
	}
	return s, nil
}

// Provider is an instance provider (such as Packet)
type Provider struct {
	provider pb.Provider
}

func (p *Provider) String() (string, error) {
	return p.provider.String(), nil
}

// NewInstance returns a new instance
func NewInstance(consulClient *consul.Client, id string) (*Instance, error) {
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
	services, err := i.Services.ToMessage()
	if err != nil {
		return nil, err
	}
	return &pb.Instance{
		Id:       i.ID,
		Owner:    i.Owner,
		Provider: i.Provider.provider,
		Device:   i.Device,
		Services: services,
	}, nil
}

// GetInstance gets instance info
func (i *Instance) GetInstance(consulClient *consul.Client) (*Instance, error) {
	kv := consulClient.KV()
	kvs, _, err := kv.List(i.instancePrefix(), nil)
	if err != nil {
		return nil, err
	}

	m, err := consulkvjson.ConsulKVsToJSON(kvs)
	if err != nil {
		return nil, err
	}

	marshalledJSON, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	owner, dataType, _, err := jsonparser.Get(marshalledJSON, "instances", i.ID, "owner")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		owner = nil
	}

	device, dataType, _, err := jsonparser.Get(marshalledJSON, "instances", i.ID, "device")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		device = nil
	}

	provider, dataType, _, err := jsonparser.Get(marshalledJSON, "instances", i.ID, "provider")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		provider = nil
	}

	serviceList := make([]*Service, 0)
	services, dataType, _, _ := jsonparser.Get(marshalledJSON, "instances", i.ID, "services")
	if dataType == jsonparser.NotExist {
		services = nil
	} else {
		jsonparser.ObjectEach(services, func(service, config []byte, dataType jsonparser.ValueType, offset int) error {
			log.Printf("%s", service)
			serviceList = append(serviceList, &Service{
				Type:   string(service),
				Config: string(config),
			})
			return nil
		})
	}

	p, err := NewProvider(string(provider))
	if err != nil {
		return nil, err
	}

	i.Provider = p
	i.Owner = string(owner)
	i.Device = string(device)
	i.Services = serviceList

	return i, nil
}

// CreateInstance creates the key/value pairs for a new instance in Consul
func CreateInstance(consulClient *consul.Client, instanceParams CreateInstanceRequest) (*Instance, error) {
	kv := consulClient.KV()

	ops := consul.KVTxnOps{
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

// DestroyInstance removes an instance from Consul
func (i *Instance) DestroyInstance(consulClient *consul.Client) error {
	kv := consulClient.KV()
	// acl := consulClient.ACL()

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + i.ID + "/",
		},
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return err
	}

	// TODO delete corresponding ACL from consul

	if !ok {
		return errors.New("Could not remove instance in Consul")
	}

	return nil
}

// SetInstanceFields sets instance/instanceID/fieldName to fieldValue
func (i *Instance) SetInstanceFields(consulClient *consul.Client, instanceFields map[string]string) (*Instance, error) {
	// TODO add some sanity checks - only allow certain fields to be set?
	// ensure that instance exists first?
	kv := consulClient.KV()

	ops := consul.KVTxnOps{}

	for field, value := range instanceFields {
		ops = append(ops, &consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + i.ID + "/" + field,
			Value: []byte(value),
		})
	}

	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not set fields in consul")
	}

	instance, err := i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// AddService adds a service in consul
func AddService(consulClient *consul.Client, instanceID, service, config string) (*Instance, error) {
	kv := consulClient.KV()

	// throw error if service already exists
	s, _ := GetService(consulClient, instanceID, service)
	if s != nil {
		return nil, errors.New("service already exists")
	}

	// TODO: add a check to handle case when config is empty object. Right now, if there's no config, no service is created.
	kvs, err := consulkvjson.ToKVs([]byte(config))
	if err != nil {
		return nil, err
	}

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + instanceID + "/services/" + service,
		},
	}
	for _, kv := range kvs {
		ops = append(ops, &consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + instanceID + "/services/" + service + "/" + kv.Key,
			Value: []byte(kv.Value),
		})
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not set service config")
	}

	i := Instance{ID: instanceID}
	instance, err := i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// GetService returns the service requested
func GetService(consulClient *consul.Client, instanceID, serviceType string) (*Service, error) {
	kv := consulClient.KV()
	serviceKVPairs, _, err := kv.List("instances/"+instanceID+"/services/"+serviceType, nil)
	if err != nil {
		return nil, err
	}
	if len(serviceKVPairs) == 0 {
		return nil, errors.New("service not found")
	}
	serviceJSON, err := consulkvjson.ConsulKVsToJSON(serviceKVPairs)
	if err != nil {
		return nil, err
	}
	serviceJSONMarshalled, err := json.Marshal(serviceJSON)
	config, dataType, _, err := jsonparser.Get(serviceJSONMarshalled, "instances", instanceID, "services", serviceType)
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		return nil, errors.New("could not retrieve service config")
	}

	return &Service{
		Type:   serviceType,
		Config: string(config),
	}, nil
}

// ConfigureService sets the configuration for a service in Consul
func ConfigureService(consulClient *consul.Client, instanceID, serviceType, config string) (*Service, error) {
	kv := consulClient.KV()
	s, err := GetService(consulClient, instanceID, serviceType)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, errors.New("problem with service")
	}
	// TODO: add a check to handle case when config is empty object. Right now, if there's no config, no service is created.
	kvs, err := consulkvjson.ToKVs([]byte(config))
	if err != nil {
		return nil, err
	}

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + instanceID + "/services/" + serviceType,
		},
	}
	for _, kv := range kvs {
		ops = append(ops, &consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + instanceID + "/services/" + serviceType + "/" + kv.Key,
			Value: []byte(kv.Value),
		})
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("could not configure service")
	}

	service, err := GetService(consulClient, instanceID, serviceType)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// RemoveService removes a service from Consul
func RemoveService(consulClient *consul.Client, instanceID, service string) (*Instance, error) {
	kv := consulClient.KV()

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + instanceID + "/services/" + service,
		},
	}

	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not remove service")
	}

	i := Instance{ID: instanceID}
	instance, err := i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
