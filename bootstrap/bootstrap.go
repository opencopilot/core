package bootstrap

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	consul "github.com/hashicorp/consul/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/julienschmidt/httprouter"
	instance "github.com/opencopilot/core/instance"
	packet "github.com/packethost/packngo"
)

// Bootstrap
type Bootstrap struct {
	ConsulCli   *consul.Client
	VaultCli    *vault.Client
	BindAddress string
	TLSCert     string
	TLSKey      string
	Payload     map[string]interface{}
}

func (b *Bootstrap) handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	instanceID := ps.ByName("instanceId")
	authPayload := r.Header.Get("Authorization")

	i, err := instance.NewInstance(b.ConsulCli, instanceID)
	if err != nil {
		http.Error(w, "Problem getting instance", 500)
		return
	}

	clientAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Could not parse client IP", 500)
		return
	}

	clientIP := net.ParseIP(clientAddr)

	verified, err := verify(b.ConsulCli, i, clientIP, authPayload)
	if err != nil || !verified {
		http.Error(w, "Could not verify device", 500)
		return
	}

	t := b.VaultCli.Auth().Token()
	bootstrapToken, err := t.Create(&vault.TokenCreateRequest{
		Policies: []string{"bootstrap"},
	})
	if err != nil {
		http.Error(w, "Could not issue bootstrap token", 500)
		return
	}

	payload := b.Payload
	payload["bootstrap_token"] = bootstrapToken

	json.NewEncoder(w).Encode(map[string]interface{}{
		"instance": instanceID,
		"payload":  payload,
	})
}

func verify(consulCli *consul.Client, i *instance.Instance, clientAddr net.IP, authPayload string) (bool, error) {
	provider, err := i.Provider.String()
	if err != nil {
		return false, errors.New("Problem handling Provider")
	}
	switch provider {
	case "PACKET":
		packetClient := packet.NewClientWithAuth("", authPayload, nil)
		device, _, err := packetClient.Devices.Get(i.Device)
		if err != nil {
			return false, err
		}
		for _, ip := range device.Network {
			deviceIP := net.ParseIP(ip.Address)
			if deviceIP == nil {
				continue
			}
			if ip.Management && clientAddr.Equal(deviceIP) {
				return true, nil
			}
		}
	default:
		return false, errors.New("Invalid provider specified")
	}
	return false, errors.New("This should be unreachable")
}

// Serve runs the http bootstrap server
func (b *Bootstrap) Serve() {
	router := httprouter.New()
	router.GET("/bootstrap/:instanceId", b.handler)
	log.Fatal(http.ListenAndServeTLS(b.BindAddress, b.TLSCert, b.TLSKey, router))
}
