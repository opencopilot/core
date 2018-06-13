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
	boostrap "github.com/opencopilot/core/bootstrap"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	pb "github.com/opencopilot/core/core"
	pbHealth "github.com/opencopilot/core/health"
)

var (
	// ConsulEncrypt is the encryption key for consul
	ConsulEncrypt = os.Getenv("CONSUL_ENCRYPT")
	// BindAddress is the interface and port the core should bind to for gRPC
	BindAddress = os.Getenv("BIND_ADDRESS")
	// BootstrapBindAddress is the interface and port the core should bind to for the HTTPS bootstrap server
	BootstrapBindAddress = os.Getenv("BOOTSTRAP_BIND_ADDRESS")
	// TLSDirectory is the path to the directory holding certs/key for TLS with consul
	TLSDirectory = os.Getenv("TLS_DIRECTORY")
	// PublicAddress is where core can be reached
	PublicAddress = os.Getenv("PUBLIC_ADDRESS")
)

func startGRPC(consulCli *consul.Client, vaultCli *vault.Client) {
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
		vaultClient:  vaultCli,
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
			Interval: "20s",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	consulClientConfig := consul.DefaultConfig()

	if os.Getenv("CONSUL_ADDRESS") != "" {
		consulClientConfig.Address = os.Getenv("CONSUL_ADDRESS")
	}

	// this should probably be stored in Vault
	if ConsulEncrypt == "" {
		log.Fatalf("CONSUL_ENCRYPT env not provided")
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if vaultToken == "" {
		log.Fatalf("VAULT_TOKEN env not provided")
	}

	if BindAddress == "" {
		BindAddress = "0.0.0.0:50060"
	}

	if BootstrapBindAddress == "" {
		BootstrapBindAddress = "0.0.0.0:5000"
	}

	if TLSDirectory == "" {
		TLSDirectory = "/opt/consul/tls/"
	}

	vaultCA := "/opt/vault/tls/vault-ca.crt"

	if os.Getenv("VAULT_CA") != "" {
		vaultCA = os.Getenv("VAULT_CA")
	}

	bootstrapCert := os.Getenv("BOOTSTRAP_CERT")
	bootstrapKey := os.Getenv("BOOTSTRAP_KEY")

	if bootstrapCert == "" || bootstrapKey == "" {
		log.Fatalf("bootstrap TLS cert or key not provided")
	}

	consulCli, err := consul.NewClient(consulClientConfig)
	if err != nil {
		log.Fatalf("failed to setup consul client: %v", err)
	}

	vaultClientConfig := vault.DefaultConfig()
	err = vaultClientConfig.ConfigureTLS(&vault.TLSConfig{
		CACert: vaultCA,
	})

	if err != nil {
		log.Fatalf("failed to configure vault client: %v", err)
	}

	vaultCli, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		log.Fatalf("failed to setup vault client: %v", err)
	}
	vaultCli.SetToken(vaultToken)

	registerCoreService(consulCli)

	log.Println("starting core...")
	go startGRPC(consulCli, vaultCli)

	log.Println("starting bootstrap HTTP server")
	b := &boostrap.Bootstrap{
		ConsulCli: consulCli,
		VaultCli:  vaultCli,
		Payload: map[string]interface{}{
			"consul_encrypt": ConsulEncrypt,
		},
		TLSCert:     bootstrapCert,
		TLSKey:      bootstrapKey,
		BindAddress: BootstrapBindAddress,
	}
	b.Serve()
}
