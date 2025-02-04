package settings

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/qdm12/gluetun/internal/configuration/settings/helpers"
	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gotree"
)

// PortForwarding contains settings for port forwarding.
type PortForwarding struct {
	// Enabled is true if port forwarding should be activated.
	// It cannot be nil for the internal state.
	Enabled *bool
	// Filepath is the port forwarding status file path
	// to use. It can be the empty string to indicate not
	// to write to a file. It cannot be nil for the
	// internal state
	Filepath *string
}

func (p PortForwarding) validate(vpnProvider string) (err error) {
	if !*p.Enabled {
		return nil
	}

	// Validate Enabled
	validProviders := []string{providers.PrivateInternetAccess}
	if !helpers.IsOneOf(vpnProvider, validProviders...) {
		return fmt.Errorf("%w: for provider %s, it is only available for %s",
			ErrPortForwardingEnabled, vpnProvider, strings.Join(validProviders, ", "))
	}

	// Validate Filepath
	if *p.Filepath != "" { // optional
		_, err := filepath.Abs(*p.Filepath)
		if err != nil {
			return fmt.Errorf("filepath is not valid: %w", err)
		}
	}

	return nil
}

func (p *PortForwarding) copy() (copied PortForwarding) {
	return PortForwarding{
		Enabled:  helpers.CopyPointer(p.Enabled),
		Filepath: helpers.CopyPointer(p.Filepath),
	}
}

func (p *PortForwarding) mergeWith(other PortForwarding) {
	p.Enabled = helpers.MergeWithPointer(p.Enabled, other.Enabled)
	p.Filepath = helpers.MergeWithPointer(p.Filepath, other.Filepath)
}

func (p *PortForwarding) overrideWith(other PortForwarding) {
	p.Enabled = helpers.OverrideWithPointer(p.Enabled, other.Enabled)
	p.Filepath = helpers.OverrideWithPointer(p.Filepath, other.Filepath)
}

func (p *PortForwarding) setDefaults() {
	p.Enabled = helpers.DefaultPointer(p.Enabled, false)
	p.Filepath = helpers.DefaultPointer(p.Filepath, "/tmp/gluetun/forwarded_port")
}

func (p PortForwarding) String() string {
	return p.toLinesNode().String()
}

func (p PortForwarding) toLinesNode() (node *gotree.Node) {
	if !*p.Enabled {
		return nil
	}

	node = gotree.New("Automatic port forwarding settings:")
	node.Appendf("Enabled: yes")

	filepath := *p.Filepath
	if filepath == "" {
		filepath = "[not set]"
	}
	node.Appendf("Forwarded port file path: %s", filepath)

	return node
}
