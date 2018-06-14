package application

import (
	"encoding/json"
	"errors"

	"github.com/buger/jsonparser"
	consul "github.com/hashicorp/consul/api"
	"github.com/opencopilot/consulkvjson"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	"github.com/opencopilot/core/provider"
	"github.com/opencopilot/core/service"
)

// Application is a set of instances
type Application struct {
	ID        string
	Type      string
	Owner     string
	Provider  *provider.Provider
	Services  service.Services
	Instances []*instance.Instance
}

// CreateApplicationRequest is a request to create a new application
type CreateApplicationRequest struct {
	ID       string
	Provider string
	Owner    string
	Type     string
}

// ToMessage serializes an application
func (a *Application) ToMessage() (*pb.Application, error) {
	services, err := a.Services.ToMessage()
	if err != nil {
		return nil, err
	}
	// TODO: move this to the instance package? Similar to how serializing services is handled above?
	instances := make([]*pb.Instance, 0)
	for _, inst := range a.Instances {
		serialized, err := inst.ToMessage()
		if err != nil {
			return nil, err
		}
		instances = append(instances, serialized)
	}
	return &pb.Application{
		Id:        a.ID,
		Type:      a.Type,
		Owner:     a.Owner,
		Provider:  a.Provider.PbProvider,
		Services:  services,
		Instances: instances,
	}, nil
}

// NewApplication creates a new Application
func NewApplication(consulClient *consul.Client, id string) (*Application, error) {
	if id == "" {
		return nil, errors.New("No application ID specified")
	}
	a := Application{
		ID: id,
	}
	application, err := a.GetApplication(consulClient)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (a *Application) applicationPrefix() string {
	return "applications/" + a.ID
}

func (a *Application) servicesPrefix() string {
	return "applications/" + a.ID + "/services/"
}

// GetApplication gets an application
func (a *Application) GetApplication(consulClient *consul.Client) (*Application, error) {
	kv := consulClient.KV()
	kvs, _, err := kv.List(a.applicationPrefix(), nil)
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

	owner, dataType, _, err := jsonparser.Get(marshalledJSON, "applications", a.ID, "owner")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		owner = nil
	}

	appType, dataType, _, err := jsonparser.Get(marshalledJSON, "applications", a.ID, "type")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		appType = nil
	}

	prov, dataType, _, err := jsonparser.Get(marshalledJSON, "applications", a.ID, "provider")
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		prov = nil
	}

	serviceList := make([]*service.Service, 0)
	services, dataType, _, _ := jsonparser.Get(marshalledJSON, "applications", a.ID, "services")
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

	instanceList := make([]*instance.Instance, 0)
	instances, dataType, _, _ := jsonparser.Get(marshalledJSON, "applications", a.ID, "instances")
	if dataType == jsonparser.NotExist {
		instances = nil
	} else {
		jsonparser.ObjectEach(instances, func(i, config []byte, dataType jsonparser.ValueType, offset int) error {
			inst, err := instance.NewInstance(consulClient, string(i))
			if err != nil {
				return err
			}
			inst, err = inst.GetInstance(consulClient)
			if err != nil {
				return err
			}
			instanceList = append(instanceList, inst)
			return nil
		})
	}

	p, err := provider.NewProvider(string(prov))
	if err != nil {
		return nil, err
	}

	a.Provider = p
	a.Owner = string(owner)
	a.Type = string(appType)
	a.Services = serviceList
	a.Instances = instanceList

	return a, nil
}

// CreateApplication creates an application
func CreateApplication(consulClient *consul.Client, applicationParams *CreateApplicationRequest) (*Application, error) {
	kv := consulClient.KV()

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "applications/" + applicationParams.ID + "/provider",
			Value: []byte(applicationParams.Provider),
		},
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "applications/" + applicationParams.ID + "/owner",
			Value: []byte(applicationParams.Owner),
		},
		&consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "applications/" + applicationParams.ID + "/type",
			Value: []byte(applicationParams.Type),
		},
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not create application in Consul")
	}

	a := Application{ID: applicationParams.ID}
	application, err := a.GetApplication(consulClient)
	if err != nil {
		return nil, err
	}

	return application, nil
}

// DestroyApplication destroys an application in Consul
func (a *Application) DestroyApplication(consulClient *consul.Client) error {
	kv := consulClient.KV()
	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "applications/" + a.ID + "/",
		},
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Could not remove application in Consul")
	}
	return nil
}

// AddInstance adds an instance to the application
func (a *Application) AddInstance(consulClient *consul.Client, i *instance.Instance) (*Application, error) {
	kv := consulClient.KV()

	_, err := kv.Put(&consul.KVPair{Key: "applications/" + a.ID + "/instances/" + i.ID}, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// RemoveInstance removes an instance from the application
func (a *Application) RemoveInstance(consulClient *consul.Client, i *instance.Instance) (*Application, error) {
	kv := consulClient.KV()

	_, err := kv.Delete("applications/"+a.ID+"/instances/"+i.ID, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}
