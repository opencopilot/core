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
)

const port = ":50060"

var (
	// ConsulEncrypt is the encryption key for consul
	ConsulEncrypt = os.Getenv("CONSUL_ENCRYPT")
)

func startGRPC(consulCli *consul.Client) {
	lis, err := net.Listen("tcp", port)
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

	pb.RegisterCoreServer(s, &server{
		consulClient: consulCli,
	})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	s.Serve(lis)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
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

	consulCli, err := consul.NewClient(consulClientConfig)
	if err != nil {
		log.Fatalf("failed to setup consul client on gRPC server")
	}

	log.Println("Starting core...")
	go startGRPC(consulCli)

	log.Println("Starting bootstrap HTTP server")
	bootstrap.Serve(consulCli, map[string]interface{}{
		"consul_encrypt": ConsulEncrypt,
	})
}
