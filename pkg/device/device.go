package device

// packeg is only used for buzzer-ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Device represents a device with name an IP
type Device struct {
	Name string
	IP   string
}

// Devices represents a list of Device
type Devices struct {
	sync.Mutex
	Items map[string]Device
}

// NewDevices returns a new Devices
func NewDevices() *Devices {
	return &Devices{Items: make(map[string]Device)}
}

// Register device name on URL
func Register(name, url string) error {

	ifAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	addrs := []string{}
	for _, a := range ifAddrs {
		addrs = append(addrs, a.String())
	}

	//
	dev := Device{
		Name: name,
		IP:   strings.Join(addrs, ", "),
	}
	body, err := json.Marshal(dev)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return nil
}
