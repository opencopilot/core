package instance

import (
	"encoding/json"
	"errors"

	"github.com/buger/jsonparser"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/opencopilot/consulkvjson"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/provider"
	service "github.com/opencopilot/core/service"
)

// Instance is a open-copilot managed instance
type Instance struct {
	ID       string
	Provider *provider.Provider
	Services service.Services
	Owner    string
	Device   string
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
		Provider: i.Provider.PbProvider,
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

	prov, dataType, _, err := jsonparser.Get(marshalledJSON, "instances", i.ID, "provider")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		prov = nil
	}

	serviceList := make([]*service.Service, 0)
	services, dataType, _, _ := jsonparser.Get(marshalledJSON, "instances", i.ID, "services")
	if dataType == jsonparser.NotExist {
		services = nil
	} else {
		jsonparser.ObjectEach(services, func(s, config []byte, dataType jsonparser.ValueType, offset int) error {
			serviceList = append(serviceList, &service.Service{
				Type:   string(s),
				Config: string(config),
			})
			return nil
		})
	}

	p, err := provider.NewProvider(string(prov))
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
func CreateInstance(consulClient *consul.Client, vaultClient *vault.Client, instanceParams CreateInstanceRequest) (*Instance, error) {
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
func (i *Instance) DestroyInstance(consulClient *consul.Client, vaultClient *vault.Client) error {
	kv := consulClient.KV()
	acl := consulClient.ACL()
	logical := vaultClient.Logical()

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

	if !ok {
		return errors.New("Could not remove instance in Consul")
	}

	tokens, _, err := acl.List(nil)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		if token.Name == "instance-"+i.ID {
			_, err = acl.Destroy(token.ID, nil)
			if err != nil {
				return err
			}
		}
	}

	_, err = logical.Delete("secret/bootstrap/" + i.ID)
	if err != nil {
		return err
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
func (i *Instance) AddService(consulClient *consul.Client, service, config string) (*Instance, error) {
	kv := consulClient.KV()
	instanceID := i.ID

	// throw error if service already exists
	s, _ := i.GetService(consulClient, service)
	if s != nil {
		return nil, errors.New("service already exists")
	}

	// TODO: add a check to handle case when config is empty object. Right now, if there's no config, no service is created.
	// i.e. there should be an initial config for each service
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

	i, err = i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// GetService returns the service requested
func (i *Instance) GetService(consulClient *consul.Client, serviceType string) (*service.Service, error) {
	kv := consulClient.KV()
	instanceID := i.ID
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

	return &service.Service{
		Type:   serviceType,
		Config: string(config),
	}, nil
}

// ConfigureService sets the configuration for a service in Consul
func (i *Instance) ConfigureService(consulClient *consul.Client, serviceType, config string) (*service.Service, error) {
	kv := consulClient.KV()
	instanceID := i.ID

	s, err := i.GetService(consulClient, serviceType)
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

	service, err := i.GetService(consulClient, serviceType)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// RemoveService removes a service from Consul
func (i *Instance) RemoveService(consulClient *consul.Client, service string) (*Instance, error) {
	kv := consulClient.KV()
	instanceID := i.ID

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

	i, err = i.GetInstance(consulClient)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// GenerateConsulToken generates an ACL in consul for this instance
func (i *Instance) GenerateConsulToken(consulClient *consul.Client) (string, error) {
	acl := consulClient.ACL()
	token, _, err := acl.Create(&consul.ACLEntry{
		Name: "instance-" + i.ID,
		Type: consul.ACLClientType,
		// TODO: move this out to a file template?
		Rules: `key "instances/` + i.ID + `" { policy = "read" }
node "` + i.ID + `" { policy = "write" }
service "opencopilot-agent" { policy = "write" }
`,
	}, nil)
	if err != nil {
		return "", err
	}

	return token, nil
}
