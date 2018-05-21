package main

import (
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	packet "github.com/packethost/packngo"
)

// PacketAuth implements auth for Packet instances
type PacketAuth struct {
	*pb.Auth
}

// Verify that a given Auth payload has access to a Packet account
func (p *PacketAuth) Verify() bool {
	client := packet.NewClientWithAuth("", p.Payload, nil)
	user, _, err := client.Users.Current()
	if err != nil {
		return false
	}
	return user.ID != ""
}

// CanManageInstance verifies that the passed in authentication can manage the specified instance, if it's a Packet instance
func (p *PacketAuth) CanManageInstance(instance *instance.Instance) bool {
	client := packet.NewClientWithAuth("", p.Payload, nil)
	device, _, err := client.Devices.Get(instance.Device)
	if err != nil {
		return false
	}
	if device != nil {
		return true
	}
	return false
}

// VerifyAuthentication verifies that a given Auth payload can authenticate to the provider it specifies
func VerifyAuthentication(auth *pb.Auth) bool {
	switch auth.Provider {
	case pb.Provider_PACKET:
		authProvider := PacketAuth{auth}
		return authProvider.Verify()
	default:
		return false
	}
}

// CanManageInstance checks whether or not the passed in authentication can manage the specified instance
func CanManageInstance(auth *pb.Auth, instance *instance.Instance) bool {
	switch auth.Provider {
	case pb.Provider_PACKET:
		authProvider := PacketAuth{auth}
		return authProvider.CanManageInstance(instance)
	default:
		return false
	}
}
