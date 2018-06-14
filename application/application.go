package application

import (
	"encoding/json"
	"errors"
	"log"

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
	return &pb.Application{
		Id:       a.ID,
		Owner:    a.Owner,
		Provider: a.Provider.PbProvider,
		Services: services,
	}, nil
}

// NewApplication creates a new Application
func NewApplication(consulClient *consul.Client, id string) (*Application, error) {
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
	log.Println(owner)
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		owner = nil
	}

	appType, dataType, _, err := jsonparser.Get(marshalledJSON, "applications", a.ID, "type")
	log.Println(appType)
	if err != nil {
		return nil, err
	}
	if dataType == jsonparser.NotExist {
		appType = nil
	}

	prov, dataType, _, err := jsonparser.Get(marshalledJSON, "applications", a.ID, "provider")
	log.Println(prov)
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

	p, err := provider.NewProvider(string(prov))
	if err != nil {
		return nil, err
	}

	a.Provider = p
	a.Owner = string(owner)
	a.Type = string(appType)
	a.Services = serviceList

	return a, nil
}

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
