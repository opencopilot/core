package service

import (
	pb "github.com/opencopilot/core/core"
)

// Service is a managed service
type Service struct {
	Type   string
	Config string
}

// ToMessage serializes a Service for gRPC
func (s *Service) ToMessage() (*pb.ServiceSpec, error) {
	return &pb.ServiceSpec{
		Type:   s.Type,
		Config: s.Config,
	}, nil
}

// Services is a list of Service
type Services []*Service

// ToMessage serializes a list of Services for gRPC
func (services Services) ToMessage() ([]*pb.ServiceSpec, error) {
	s := make([]*pb.ServiceSpec, 0)
	for _, service := range services {
		serialized, err := service.ToMessage()
		if err != nil {
			return nil, err
		}
		s = append(s, serialized)
	}
	return s, nil
}
