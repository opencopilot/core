package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	packet "github.com/packethost/packngo"
)

var (
	// PacketProjectID is the packet project id where instances should be created. Figure this out...
	PacketProjectID = os.Getenv("PACKET_PROJECT_ID")
)

// GetPacketInstance gets an instance by ID
func GetPacketInstance(consulClient *consul.Client, in *pb.GetInstanceRequest) (*pb.Instance, error) {

	i, err := instance.NewInstance(consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	instanceMessage, err := i.ToMessage()
	if err != nil {
		return nil, err
	}
	return instanceMessage, nil
}

// CreatePacketInstance creates the necessary data structures in Consul for a new instance, and provisions a device on Packet
func CreatePacketInstance(consulClient *consul.Client, in *pb.CreateInstanceRequest) (*pb.Instance, error) {
	id := uuid.New()

	packetClient := packet.NewClientWithAuth("", in.Auth.Payload, nil)
	_, _, err := packetClient.Users.Current()
	if err != nil {
		return nil, err
	}

	// TODO: figure this out...
	projID := PacketProjectID

	instance, err := instance.CreateInstance(consulClient, instance.CreateInstanceRequest{
		ID:       id.String(),
		Owner:    projID,
		Device:   "",
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

	customDataJSONString, err := json.Marshal(customData)
	if err != nil {
		return nil, err
	}

	createReq := packet.DeviceCreateRequest{
		Hostname:     "open-copilot-instance-" + id.String(),
		ProjectID:    projID,
		Facility:     "ewr1",
		Plan:         "baremetal_2",
		OS:           "ubuntu_16_04",
		BillingCycle: "hourly",
		CustomData:   string(customDataJSONString),
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

	instanceMessage, err := instance.ToMessage()
	return instanceMessage, nil
}

// DestroyPacketInstance destroys a packet instance
func DestroyPacketInstance(consulClient *consul.Client, in *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	packetClient := packet.NewClientWithAuth("", in.Auth.Payload, nil)

	instance, err := instance.NewInstance(consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	device, _, err := packetClient.Devices.Get(instance.Device)
	if err != nil {
		return nil, err
	}
	if device.State != "active" {
		return nil, errors.New("Device is still provisioning")
	}

	err = instance.DestroyInstance(consulClient)
	if err != nil {
		return nil, err
	}

	packetClient.Devices.Delete(instance.Device)

	return &pb.DestroyInstanceResponse{}, nil
}
