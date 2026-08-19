package wireproxy

import (
	"bytes"
	"fmt"
	"sync"

	"net/netip"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/amnezia-vpn/amneziawg-go/v3/conn"
	"github.com/amnezia-vpn/amneziawg-go/v3/device"
	"github.com/amnezia-vpn/amneziawg-go/v3/tun/netstack"
)

// DeviceSetting contains the parameters for setting up a tun interface
type DeviceSetting struct {
	IpcRequest string
	DNS        []netip.Addr
	DeviceAddr []netip.Addr
	MTU        int
}

// CreateIPCRequest serialize the config into an IPC request and DeviceSetting
func CreateIPCRequest(conf *DeviceConfig) (*DeviceSetting, error) {
	var request bytes.Buffer

	fmt.Fprintf(&request, "private_key=%s\n", conf.SecretKey)

	if conf.ListenPort != nil {
		fmt.Fprintf(&request, "listen_port=%d\n", *conf.ListenPort)
	}

	awgOrder := []string{
		"jc", "jmin", "jmax", "s1", "s2", "s3", "s4",
		"h1", "h2", "h3", "h4", "i1", "i2", "i3", "i4", "i5",
		"header_protection_key", "content_padding_addition", "rekey_after_time",
		"rekey_timeout", "reject_after_time", "keepalive_timeout", "max_handshake_attempts",
	}
	for _, key := range awgOrder {
		if value, ok := conf.AWGParameters[key]; ok {
			fmt.Fprintf(&request, "%s=%s\n", key, value)
		}
	}

	for _, peer := range conf.Peers {
		fmt.Fprintf(&request, heredoc.Doc(`
				public_key=%s
				persistent_keepalive_interval=%s
				preshared_key=%s
			`),
			peer.PublicKey, peer.KeepAlive, peer.PreSharedKey,
		)
		if peer.Endpoint != nil {
			fmt.Fprintf(&request, "endpoint=%s\n", *peer.Endpoint)
		}

		if len(peer.AllowedIPs) > 0 {
			for _, ip := range peer.AllowedIPs {
				fmt.Fprintf(&request, "allowed_ip=%s\n", ip.String())
			}
		} else {
			request.WriteString(heredoc.Doc(`
				allowed_ip=0.0.0.0/0
				allowed_ip=::0/0
			`))
		}
	}

	setting := &DeviceSetting{IpcRequest: request.String(), DNS: conf.DNS, DeviceAddr: conf.Endpoint, MTU: conf.MTU}
	return setting, nil
}

// StartWireguard creates a tun interface on netstack given a configuration
func StartWireguard(conf *Configuration, logLevel int) (*VirtualTun, error) {
	return StartWireguardWithLogger(conf, device.NewLogger(logLevel, ""))
}

// StartWireguardWithLogger creates a tun interface on netstack given a configuration
// and a custom device logger, so an embedding application can capture engine logs.
func StartWireguardWithLogger(conf *Configuration, logger *device.Logger) (*VirtualTun, error) {
	deviceConf := conf.Device
	setting, err := CreateIPCRequest(deviceConf)
	if err != nil {
		return nil, err
	}

	tun, tnet, err := netstack.CreateNetTUN(setting.DeviceAddr, setting.DNS, setting.MTU)
	if err != nil {
		return nil, err
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), logger)
	err = dev.IpcSet(setting.IpcRequest)
	if err != nil {
		return nil, err
	}

	err = dev.Up()
	if err != nil {
		return nil, err
	}

	hasV4 := false
	hasV6 := false
	for _, addr := range setting.DeviceAddr {
		if addr.Is4() {
			hasV4 = true
		}
		if addr.Is6() {
			hasV6 = true
		}
	}

	if conf.Resolve.ResolveStrategy == "auto" {
		if hasV4 && !hasV6 {
			conf.Resolve.ResolveStrategy = "ipv4"
		} else {
			conf.Resolve.ResolveStrategy = "ipv6"
		}
	}
	return &VirtualTun{
		Tnet:           tnet,
		Dev:            dev,
		Conf:           deviceConf,
		ResolveConfig:  conf.Resolve,
		SystemDNS:      len(setting.DNS) == 0,
		PingRecord:     make(map[string]uint64),
		PingRecordLock: new(sync.Mutex),
		life:           &lifecycle{},
	}, nil
}
