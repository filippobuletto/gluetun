package routing

import (
	"errors"
	"fmt"
	"net"
)

func IPIsPrivate(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

var (
	errInterfaceIPNotFound = errors.New("IP address not found for interface")
	errInterfaceListAddr   = errors.New("cannot list interface addresses")
	errInterfaceNotFound   = errors.New("network interface not found")
)

func (r *Routing) assignedIP(interfaceName string) (ip net.IP, err error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", errInterfaceNotFound, interfaceName, err)
	}
	addresses, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", errInterfaceListAddr, interfaceName, err)
	}
	for _, address := range addresses {
		switch value := address.(type) {
		case *net.IPAddr:
			return value.IP, nil
		case *net.IPNet:
			return value.IP, nil
		}
	}
	return nil, fmt.Errorf("%w: interface %s in %d addresses",
		errInterfaceIPNotFound, interfaceName, len(addresses))
}