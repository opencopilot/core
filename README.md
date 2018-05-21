
![OpenCoPilot Logo](docs/img/OC_transparent.png)

## OpenCoPilot Core

This is a `gRPC` server that speaks to a `Consul` cluster and the Packet API.

### Generate from gRPC .proto Definitions

`protoc -I ./core ./core/Core.proto --go_out=plugins=grpc:./core`

`protoc -I ./agent ./agent/Agent.proto --go_out=plugins=grpc:./agent`