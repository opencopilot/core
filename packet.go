package main

import (
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
func GetPacketInstance(consulClient consul.Client, in *pb.GetInstanceRequest) (*pb.Instance, error) {

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
func CreatePacketInstance(consulClient consul.Client, in *pb.CreateInstanceRequest) (*pb.Instance, error) {
	id := uuid.New()

	packetClient := packet.NewClient("", in.Auth.Payload, nil)
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

	createReq := packet.DeviceCreateRequest{
		Hostname:     "test-provision-" + id.String(),
		ProjectID:    projID,
		Facility:     "ewr1",
		Plan:         "baremetal_2",
		OS:           "ubuntu_16_04",
		BillingCycle: "hourly",
	}
	device, _, err := packetClient.Devices.Create(&createReq)
	if err != nil {
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
func DestroyPacketInstance(consulClient consul.Client, in *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	kv := consulClient.KV()
	packetClient := packet.NewClient("", in.Auth.Payload, nil)

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

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + in.InstanceId + "/",
		},
	}
	ok, _, _, err := kv.Txn(ops, nil)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("Could not remove instance in Consul")
	}

	packetClient.Devices.Delete(instance.Device)

	return &pb.DestroyInstanceResponse{}, nil
}
