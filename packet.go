package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	packet "github.com/packethost/packngo"
)

// GetPacketProjectFromAuthPayload returns the Packet project of a project level API key
func GetPacketProjectFromAuthPayload(auth string) (string, error) {
	packetClient := packet.NewClientWithAuth("", auth, nil)
	var project map[string]interface{}
	_, err := packetClient.DoRequest("GET", "/project", "", &project)
	if err != nil {
		return "", err
	}
	projectID, ok := project["id"]
	if !ok {
		return "", errors.New("problem verifying project from auth")
	}
	return projectID.(string), nil
}

// GetPacketInstance gets an instance by ID
func GetPacketInstance(consulClient *consul.Client, instanceID string) (*instance.Instance, error) {

	i, err := instance.NewInstance(consulClient, instanceID)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// CreatePacketInstance creates the necessary data structures in Consul for a new instance, and provisions a device on Packet
func CreatePacketInstance(consulClient *consul.Client, in *pb.CreateInstanceRequest) (*instance.Instance, error) {
	id := uuid.New()

	packetClient := packet.NewClientWithAuth("", in.Auth.Payload, nil)

	projID, err := GetPacketProjectFromAuthPayload(in.Auth.Payload)
	if err != nil {
		return nil, err
	}

	instance, err := instance.CreateInstance(consulClient, instance.CreateInstanceRequest{
		ID:       id.String(),
		Owner:    projID,
		Device:   "", // can't set this yet because we don't know what the device ID is until it's provisioned
		Provider: "PACKET",
	})
	if err != nil {
		return nil, err
	}

	acl := consulClient.ACL()
	token, _, err := acl.Create(&consul.ACLEntry{
		Name:  "instance-" + id.String(),
		Type:  consul.ACLClientType,
		Rules: `key "instances/` + id.String() + `" { policy = "read" }`,
	}, nil)
	if err != nil {
		return nil, err
	}

	customData := map[string]interface{}{
		"COPILOT": map[string]interface{}{
			"INSTANCE_ID":  id.String(),
			"CONSUL_TOKEN": token,
			// "CONSUL_ENCRYPT": ConsulEncrypt,
		},
	}

	customDataJSON, err := json.Marshal(customData)
	if err != nil {
		return nil, err
	}

	userDataString, err := ioutil.ReadFile("./assets/packet.userdata.sh")
	if err != nil {
		return nil, err
	}

	createReq := packet.DeviceCreateRequest{
		Hostname:     "open-copilot-instance-" + id.String(),
		ProjectID:    projID,
		Facility:     "ewr1",
		Plan:         "baremetal_1",
		OS:           "ubuntu_16_04",
		BillingCycle: "hourly",
		CustomData:   string(customDataJSON),
		UserData:     string(userDataString),
	}
	device, _, err := packetClient.Devices.Create(&createReq)
	if err != nil {
		instance.DestroyInstance(consulClient)
		// TODO revoke token
		return nil, err
	}

	_, err = instance.SetInstanceFields(consulClient, map[string]string{
		"device": device.ID,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// DestroyPacketInstance destroys a packet instance
func DestroyPacketInstance(consulClient *consul.Client, in *pb.DestroyInstanceRequest) error {
	packetClient := packet.NewClientWithAuth("", in.Auth.Payload, nil)

	instance, err := instance.NewInstance(consulClient, in.InstanceId)
	if err != nil {
		return err
	}

	device, _, err := packetClient.Devices.Get(instance.Device)
	if err != nil {
		return err
	}
	if device.State != "active" {
		return errors.New("Device is still provisioning")
	}

	err = instance.DestroyInstance(consulClient)
	if err != nil {
		return err
	}

	packetClient.Devices.Delete(instance.Device)

	return nil
}
