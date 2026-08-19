package main

import (
	"fmt"
	"net"
)

// tcpPortAvailable проверяет, что порт свободен на нужном интерфейсе.
func tcpPortAvailable(port int, listenOnLAN bool) bool {
	host := "127.0.0.1"
	if listenOnLAN {
		host = "0.0.0.0"
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// lanIPv4Addresses возвращает не-loopback, не-link-local IPv4 адреса машины.
func lanIPv4Addresses() []string {
	var out []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return out
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip4 := ipNet.IP.To4()
		if ip4 == nil || ip4.IsLinkLocalUnicast() {
			continue
		}
		out = append(out, ip4.String())
	}
	return out
}
