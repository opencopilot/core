#!/bin/sh
set -x

#### Install Docker ###
curl -fsSL get.docker.com -o get-docker.sh
sh get-docker.sh

mkdir /etc/consul
mkdir /opt/consul
mkdir /opt/consul/tls

COPILOT_CORE_ADDR=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.CORE_ADDR)
PACKET_AUTH=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.PACKET_AUTH)
INSTANCE_ID=$(curl -sS metadata.packet.net/metadata | jq -r .customdata.COPILOT.INSTANCE_ID)
FACILITY=$(curl -sS metadata.packet.net/metadata | jq -r .facility)
CONSUL_TLS_DIR=/opt/consul/tls

BOOTSTRAP_TOKEN=$(curl -sS -k -H "Authorization: $PACKET_AUTH" https://$COPILOT_CORE_ADDR:5000/bootstrap/$INSTANCE_ID | jq -r .bootstrap_token)
CONSUL_ENCRYPT=$(curl -sS -k -H "Authorization: $PACKET_AUTH" https://$COPILOT_CORE_ADDR:5000/bootstrap/$INSTANCE_ID | jq -r .consul_encrypt)
curl -sS -k --header "X-Vault-Token: $BOOTSTRAP_TOKEN" -H "Content-Type: application/json" -d "{\"common_name\": \"$INSTANCE_ID.opencopilot.com\", \"ttl\": \"7200h\"}" https://$COPILOT_CORE_ADDR:8200/v1/pki_consul/issue/instance_consul_tls >> $CONSUL_TLS_DIR/consul_tls.json
CONSUL_TOKEN=$(curl -sS -k --header "X-Vault-Token: $BOOTSTRAP_TOKEN" -H "Content-Type: application/json" https://$COPILOT_CORE_ADDR:8200/v1/secret/bootstrap/$INSTANCE_ID | jq -r .data.consul_token)
cat $CONSUL_TLS/consul_tls.json | jq -r .data.issuing_ca >> $CONSUL_TLS_DIR/consul-ca.crt
cat $CONSUL_TLS/consul_tls.json | jq -r .data.certificate >> $CONSUL_TLS_DIR/consul.crt
cat $CONSUL_TLS/consul_tls.json | jq -r .data.private_key >> $CONSUL_TLS_DIR/consul.key


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
        "$COPILOT_CORE_ADDR"
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
    -P \
    quay.io/opencopilot/agent
