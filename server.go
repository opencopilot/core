package main

import (
	"context"
	"errors"

	consul "github.com/hashicorp/consul/api"
	pb "github.com/opencopilot/core/core"
	"github.com/opencopilot/core/instance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	consulClient *consul.Client
}

func (s *server) GetInstance(ctx context.Context, in *pb.GetInstanceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := GetPacketInstance(s.consulClient, in)
	return instance, err
}

func (s *server) CreateInstance(ctx context.Context, in *pb.CreateInstanceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := CreatePacketInstance(s.consulClient, in)
	return instance, err
}

func (s *server) DestroyInstance(ctx context.Context, in *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	res, err := DestroyPacketInstance(s.consulClient, in)
	return res, err
}

func (s *server) AddService(ctx context.Context, in *pb.AddServiceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	i, err := instance.NewInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	i, err = i.AddService(s.consulClient, in.Service.Type, in.Service.Config)
	if err != nil {
		return nil, err
	}

	instanceMessage, err := i.ToMessage()
	if err != nil {
		return nil, err
	}

	return instanceMessage, nil
}

func (s *server) GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.Service, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	i, err := instance.NewInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	service, err := i.GetService(s.consulClient, in.ServiceType)
	if err != nil {
		return nil, err
	}

	return &pb.Service{
		Type:   service.Type,
		Config: service.Config,
	}, nil
}

func (s *server) ConfigureService(ctx context.Context, in *pb.ConfigureServiceRequest) (*pb.Service, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	i, err := instance.NewInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	service, err := i.ConfigureService(s.consulClient, in.Service.Type, in.Service.Config)
	if err != nil {
		return nil, err
	}

	return &pb.Service{
		Type:   service.Type,
		Config: service.Config,
	}, nil
}

func (s *server) RemoveService(ctx context.Context, in *pb.RemoveServiceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	i, err := instance.NewInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	i, err = i.RemoveService(s.consulClient, in.ServiceType)
	if err != nil {
		return nil, err
	}

	instanceMessage, err := i.ToMessage()
	if err != nil {
		return nil, err
	}
	return instanceMessage, nil
}
