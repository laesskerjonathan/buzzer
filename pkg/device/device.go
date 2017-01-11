package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Device struct {
	Name string
	IP   string
}

// Devices
type Devices struct {
	sync.Mutex
	Items map[string]Device
}

func NewDevices() *Devices {
	return &Devices{Items: make(map[string]Device)}
}

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
		return errors.New(fmt.Sprintf("%s: %s", url, resp.Status))
	}
	return nil
}
