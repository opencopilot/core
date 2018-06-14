package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/opencopilot/core/application"
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
func CreatePacketInstance(consulClient *consul.Client, vaultClient *vault.Client, in *pb.CreateInstanceRequest) (*instance.Instance, error) {
	id := uuid.New()

	packetClient := packet.NewClientWithAuth("", in.Auth.Payload, nil)

	projID, err := GetPacketProjectFromAuthPayload(in.Auth.Payload)
	if err != nil {
		return nil, err
	}

	instance, err := instance.CreateInstance(consulClient, vaultClient, instance.CreateInstanceRequest{
		ID:       id.String(),
		Owner:    projID,
		Device:   "", // can't set this yet because we don't know what the device ID is until it's provisioned
		Provider: "PACKET",
	})
	if err != nil {
		return nil, err
	}

	token, err := instance.GenerateConsulToken(consulClient)
	if err != nil {
		return nil, err
	}

	logical := vaultClient.Logical()
	_, err = logical.Write("secret/bootstrap/"+id.String(), map[string]interface{}{
		"consul_token": token,
	})
	if err != nil {
		return nil, err
	}

	customData := map[string]interface{}{
		"COPILOT": map[string]interface{}{
			"INSTANCE_ID": id.String(),
			"CORE_ADDR":   PublicAddress,
			"PACKET_AUTH": in.Auth.Payload,
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
		Hostname:     "opencopilot-" + strings.Split(id.String(), "-")[0],
		ProjectID:    projID,
		Facility:     in.Region,
		Plan:         in.Type,
		OS:           "ubuntu_16_04",
		BillingCycle: "hourly",
		CustomData:   string(customDataJSON),
		UserData:     string(userDataString),
	}
	device, _, err := packetClient.Devices.Create(&createReq)
	if err != nil {
		instance.DestroyInstance(consulClient, vaultClient)
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
func DestroyPacketInstance(consulClient *consul.Client, vaultClient *vault.Client, in *pb.DestroyInstanceRequest) error {
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

	err = instance.DestroyInstance(consulClient, vaultClient)
	if err != nil {
		return err
	}

	packetClient.Devices.Delete(instance.Device)

	return nil
}

// CreatePacketApplication creates an application in a Packet project
func CreatePacketApplication(consulClient *consul.Client, in *pb.CreateApplicationRequest) (*application.Application, error) {
	id := uuid.New()

	projID, err := GetPacketProjectFromAuthPayload(in.Auth.Payload)
	if err != nil {
		return nil, err
	}

	a, err := application.CreateApplication(consulClient, &application.CreateApplicationRequest{
		ID:       id.String(),
		Provider: "PACKET",
		Owner:    projID,
		Type:     in.Type,
	})

	return a, nil
}

// GetPacketApplication gets a Packet application
func GetPacketApplication(consulClient *consul.Client, applicationID string) (*application.Application, error) {

	app, err := application.NewApplication(consulClient, applicationID)
	if err != nil {
		return nil, err
	}

	return app, nil
}
