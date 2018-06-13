package main

import (
	"context"
	"errors"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	pb "github.com/opencopilot/core/core"
	pbHealth "github.com/opencopilot/core/health"
	"github.com/opencopilot/core/instance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	consulClient *consul.Client
	vaultClient  *vault.Client
}

func (s *server) Check(ctx context.Context, in *pbHealth.HealthCheckRequest) (*pbHealth.HealthCheckResponse, error) {
	return &pbHealth.HealthCheckResponse{
		Status: pbHealth.HealthCheckResponse_SERVING,
	}, nil
}

func (s *server) GetInstance(ctx context.Context, in *pb.GetInstanceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := GetPacketInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	canManage := CanManageInstance(in.Auth, instance)
	if !canManage {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	instanceMessage, err := instance.ToMessage()
	if err != nil {
		return nil, err
	}

	return instanceMessage, err
}

func (s *server) CreateInstance(ctx context.Context, in *pb.CreateInstanceRequest) (*pb.Instance, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := CreatePacketInstance(s.consulClient, s.vaultClient, in)
	if err != nil {
		return nil, err
	}

	instanceMessage, err := instance.ToMessage()
	if err != nil {
		return nil, err
	}
	return instanceMessage, err
}

func (s *server) DestroyInstance(ctx context.Context, in *pb.DestroyInstanceRequest) (*pb.EmptyResponse, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	instance, err := GetPacketInstance(s.consulClient, in.InstanceId)
	if err != nil {
		return nil, err
	}

	canManage := CanManageInstance(in.Auth, instance)
	if !canManage {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	err = DestroyPacketInstance(s.consulClient, s.vaultClient, in)
	if err != nil {
		return nil, err
	}
	return &pb.EmptyResponse{}, err
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

func (s *server) GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.ServiceSpec, error) {
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

	return &pb.ServiceSpec{
		Type:   service.Type,
		Config: service.Config,
	}, nil
}

func (s *server) ConfigureService(ctx context.Context, in *pb.ConfigureServiceRequest) (*pb.ServiceSpec, error) {
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

	return &pb.ServiceSpec{
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

func (s *server) CreateApplication(ctx context.Context, in *pb.CreateApplicationRequest) (*pb.Application, error) {
	if !VerifyAuthentication(in.Auth) {
		return nil, status.Errorf(codes.PermissionDenied, "Invalid authentication")
	}

	if in.Auth.Provider != pb.Provider_PACKET {
		return nil, errors.New("Invalid auth provider")
	}

	application, err := CreatePacketApplication(s.consulClient, in)
	if err != nil {
		return nil, err
	}

	return application.ToMessage()
}

func (s *server) DestroyApplication(ctx context.Context, in *pb.DestroyApplicationRequest) (*pb.EmptyResponse, error) {
	return nil, nil
}

func (s *server) GetApplication(ctx context.Context, in *pb.GetApplicationRequest) (*pb.Application, error) {
	return nil, nil
}

func (s *server) ApplicationAddInstance(ctx context.Context, in *pb.ApplicationAddInstanceRequest) (*pb.Application, error) {
	return nil, nil
}

func (s *server) ApplicationRemoveInstance(ctx context.Context, in *pb.ApplicationRemoveInstanceRequest) (*pb.Application, error) {
	return nil, nil
}
