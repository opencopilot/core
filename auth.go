package main

import (
	pb "github.com/opencopilot/core/core"
	packet "github.com/packethost/packngo"
)

// PacketAuth implements auth for Packet instances
type PacketAuth struct {
	*pb.Auth
}

// Verify that a given Auth payload has access to a Packet account
func (p *PacketAuth) Verify() bool {
	client := packet.NewClient("", p.Payload, nil)
	user, _, error := client.Users.Current()
	if error != nil {
		return false
	}
	return user.ID != ""
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
