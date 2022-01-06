package windscribe

import (
	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

func (w *Windscribe) GetConnection(selection settings.ServerSelection) (
	connection models.Connection, err error) {
	port := getPort(selection)
	protocol := utils.GetProtocol(selection)

	servers, err := w.filterServers(selection)
	if err != nil {
		return connection, err
	}

	var connections []models.Connection
	for _, server := range servers {
		for _, IP := range server.IPs {
			connection := models.Connection{
				Type:     selection.VPN,
				IP:       IP,
				Port:     port,
				Protocol: protocol,
				Hostname: server.OvpnX509,
				PubKey:   server.WgPubKey,
			}
			connections = append(connections, connection)
		}
	}

	return utils.PickConnection(connections, selection, w.randSource)
}

func getPort(selection settings.ServerSelection) (port uint16) {
	const (
		defaultOpenVPNTCP = 443
		defaultOpenVPNUDP = 1194
		defaultWireguard  = 1194
	)
	return utils.GetPort(selection, defaultOpenVPNTCP,
		defaultOpenVPNUDP, defaultWireguard)
}
