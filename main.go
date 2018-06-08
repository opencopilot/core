package main

import (
	"log"
	"net"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"

	consul "github.com/hashicorp/consul/api"
	"github.com/opencopilot/core/bootstrap"
	pb "github.com/opencopilot/core/core"
	pbHealth "github.com/opencopilot/core/health"
)

var (
	// ConsulEncrypt is the encryption key for consul
	ConsulEncrypt = os.Getenv("CONSUL_ENCRYPT")
	// BindAddress is the interface and port the core should bind to for gRPC
	BindAddress = os.Getenv("BIND_ADDRESS")
	// HTTPBindAddress is the interface and port the core should bind to for HTTP
	HTTPBindAddress = os.Getenv("HTTP_BIND_ADDRESS")
	// TLSDirectory is the path to the directory holding certs/key for TLS with consul
	TLSDirectory = os.Getenv("TLS_DIRECTORY")
	// PublicAddress is where core can be reached
	PublicAddress = os.Getenv("PUBLIC_ADDRESS")
)

func startGRPC(consulCli *consul.Client) {
	lis, err := net.Listen("tcp", BindAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}

	// TODO: TLS for gRPC connection to outside world
	// creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	// if err != nil {
	// 	log.Fatalf("failed to load credentials: %v", err)
	// }

	s := grpc.NewServer(
		// grpc.Creds(creds),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.StreamServerInterceptor(logger),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	coreServer := &server{
		consulClient: consulCli,
	}
	pb.RegisterCoreServer(s, coreServer)
	pbHealth.RegisterHealthServer(s, coreServer)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	s.Serve(lis)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func registerCoreService(consulCli *consul.Client) {
	agent := consulCli.Agent()
	err := agent.ServiceRegister(&consul.AgentServiceRegistration{
		Name: "opencopilot-core",
		Check: &consul.AgentServiceCheck{
			CheckID:  "core-grpc",
			Name:     "Core gRPC Health Check",
			GRPC:     BindAddress,
			Interval: "10s",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	consulClientConfig := consul.DefaultConfig()
	if os.Getenv("ENV") == "dev" {
		consulClientConfig.Address = "host.docker.internal:8500"
	}

	if os.Getenv("CONSUL_ADDRESS") != "" {
		consulClientConfig.Address = os.Getenv("CONSUL_ADDRESS")
	}

	if os.Getenv("CONSUL_TOKEN") != "" {
		consulClientConfig.Token = os.Getenv("CONSUL_TOKEN")
	}

	if ConsulEncrypt == "" {
		log.Fatalf("CONSUL_ENCRYPT env not provided")
	}

	if BindAddress == "" {
		BindAddress = "0.0.0.0:50060"
	}

	if HTTPBindAddress == "" {
		HTTPBindAddress = "0.0.0.0:5000"
	}

	if TLSDirectory == "" {
		TLSDirectory = "/opt/consul/tls/"
	}

	consulCli, err := consul.NewClient(consulClientConfig)
	if err != nil {
		log.Fatalf("failed to setup consul client on gRPC server: %v", err)
	}

	registerCoreService(consulCli)

	log.Println("Starting core...")
	go startGRPC(consulCli)

	log.Println("Starting bootstrap HTTP server")
	bootstrap.Serve(consulCli, map[string]interface{}{
		"consul_encrypt": ConsulEncrypt,
	}, HTTPBindAddress)
}
