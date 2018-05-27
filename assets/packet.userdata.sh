#!/bin/sh

#### Install Docker ###
curl -fsSL get.docker.com -o get-docker.sh
sh get-docker.sh

mkdir /etc/consul
mkdir /opt/consul
mkdir /opt/consul/tls

COPILOT_CORE_ADDR=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CORE_ADDR)
CONSUL_ENCRYPT=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_ENCRYPT)
CONSUL_TOKEN=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_TOKEN)
INSTANCE_ID=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.INSTANCE_ID)
FACILITY=$(curl -sS metadata.packet.net/metadata | jq -r .facility)
CONSUL_TLS_DIR=/opt/consul/tls

curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_CA >> $CONSUL_TLS_DIR/consul-ca.crt
curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_CERT >> $CONSUL_TLS_DIR/consul.crt
curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CONSUL_KEY >> $CONSUL_TLS_DIR/consul.key

cat > /etc/consul/config.json <<EOF
{
    "datacenter": "$FACILITY",
    "data_dir": "/opt/consul",
    "log_level": "DEBUG",
    "node_name": "$INSTANCE_ID",
    "server": false,
    "encrypt": "$CONSUL_ENCRYPT",
    "acl_datacenter": "ewr1",
    "acl_token": "$CONSUL_TOKEN",
    "retry_join": [
        "$(echo $COPILOT_CORE_ADDR)"
    ],
    "ca_file": "$CONSUL_TLS_DIR/consul-ca.crt",
    "cert_file": "$CONSUL_TLS_DIR/consul.crt",
    "key_file": "$CONSUL_TLS_DIR/consul.key"
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
    -P
    quay.io/opencopilot/agent
