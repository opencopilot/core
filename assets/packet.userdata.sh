#!/bin/sh

#### Install Docker ###
curl -fsSL get.docker.com -o get-docker.sh
sh get-docker.sh

mkdir /etc/consul
mkdir /opt/consul

# COPILOT_CORE_ADDR=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CORE_ADDR)
CONSUL_ENCRYPT=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_ENCRYPT)
CONSUL_TOKEN=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_TOKEN)
INSTANCE_ID=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.INSTANCE_ID)
FACILITY=$(curl -sS metadata.packet.net/metadata | jq -r .facility)

cat > /etc/consul/config.json <<EOF
{
    "datacenter": "$(echo $FACILITY)",
    "data_dir": "/opt/consul",
    "log_level": "DEBUG",
    "node_name": "$(echo $INSTANCE_ID)",
    "server": false,
    "encrypt": "$(echo $CONSUL_ENCRYPT)",
    "acl_datacenter": "ewr1",
    "acl_token": "$(echo CONSUL_TOKEN)",
    "verify_outgoing": true,
    "ca_file": "/opt/consul/tls/consul-ca.crt",
    "cert_file": "/opt/consul/tls/consul.crt",
    "key_file": "/opt/consul/tls/consul.key"
}
EOF

CONSUL_ADVERTISE_ADDRESS=$( curl -sS metadata.packet.net/metadata | jq -r '.network.addresses[] | select(.management == true) | select(.public == true) | select(.address_family == 4) | .address')

### Start Consul ###
docker run \
    --net="host" \
    -v /etc/consul:/etc/consul \
    -v /opt/consul:/opt/consul \
    -d \
    --restart always \
    consul agent -bind="0.0.0.0" -advertise=$CONSUL_ADVERTISE_ADDRESS -config-file="/etc/consul/config.json"

### Start Agent ###
docker run \
    -e "INSTANCE_ID=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.INSTANCE_ID)" \
    -e "CONFIG_DIR=/etc/opencopilot" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /etc/opencopilot/:/etc/opencopilot/ \
    --name agent \
    --restart always \
    -d \
    --net="host" \
    -p 50051 \
    -p 5000 \
    quay.io/opencopilot/agent
