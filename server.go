package main

import (
	"context"
	"errors"

	consul "github.com/hashicorp/consul/api"
	consulkvjson "github.com/opencopilot/consul-kv-json"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	consulClient consul.Client
}

func (s *server) GetInstance(ctx context.Context, in *pb.GetInstanceRequest) (*pb.Instance, error) {
	if authed := VerifyAuthentication(in.Auth); authed == false {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := GetPacketInstance(s.consulClient, in)
	return instance, err
}

func (s *server) CreateInstance(ctx context.Context, in *pb.CreateInstanceRequest) (*pb.Instance, error) {
	if authed := VerifyAuthentication(in.Auth); authed == false {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := CreatePacketInstance(s.consulClient, in)
	return instance, err
}

func (s *server) DestroyInstance(ctx context.Context, in *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	if authed := VerifyAuthentication(in.Auth); authed == false {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	res, err := DestroyPacketInstance(s.consulClient, in)
	return res, err
}

// TODO: this needs more thought - it's an admin only function, so what type of auth should it require?
// Does that auth permit listing instances across providers?
func (s *server) ListInstances(in *pb.ListInstancesRequest, stream pb.Core_ListInstancesServer) error {
	// if authed := VerifyAuthentication(in.Auth); authed == false {
	// 	return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	// }

	instances, err := instance.ListInstances(s.consulClient)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return nil
	}
	for _, instance := range instances {
		instanceMessage, err := instance.ToMessage()
		if err != nil {
			return err
		}
		stream.Send(instanceMessage)
	}

	return nil
}

func (s *server) AddService(ctx context.Context, in *pb.AddServiceRequest) (*pb.Instance, error) {
	if authed := VerifyAuthentication(in.Auth); authed == false {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	kv := s.consulClient.KV()

	config := in.Config

	kvs, err := consulkvjson.ToKVs([]byte(config))
	if err != nil {
		return nil, err
	}

	ops := consul.KVTxnOps{
		&consul.KVTxnOp{
			Verb: consul.KVDeleteTree,
			Key:  "instances/" + in.InstanceId + "/services/" + in.Service + "/",
		},
	}
	for _, kv := range kvs {
		ops = append(ops, &consul.KVTxnOp{
			Verb:  consul.KVSet,
			Key:   "instances/" + in.InstanceId + "/services/" + in.Service + "/" + kv.Key,
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

	i := instance.Instance{ID: in.InstanceId}
	instance, err := i.GetInstance(s.consulClient)
	if err != nil {
		return nil, err
	}

	imessage, err := instance.ToMessage()
	if err != nil {
		return nil, err
	}

	return imessage, nil
}

func (s *server) RemoveService(ctx context.Context, in *pb.RemoveServiceRequest) (*pb.Instance, error) {
	if authed := VerifyAuthentication(in.Auth); authed == false {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	return nil, nil
}
