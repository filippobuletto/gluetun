package settings

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/gluetun/internal/configuration/settings/helpers"
	"github.com/qdm12/gotree"
)

// Unbound is settings for the Unbound program.
type Unbound struct {
	Providers             []string
	Caching               *bool
	IPv6                  *bool
	VerbosityLevel        *uint8
	VerbosityDetailsLevel *uint8
	ValidationLogLevel    *uint8
	Username              string
	Allowed               []netip.Prefix
}

func (u *Unbound) setDefaults() {
	if len(u.Providers) == 0 {
		u.Providers = []string{
			provider.Cloudflare().String(),
		}
	}

	u.Caching = helpers.DefaultPointer(u.Caching, true)
	u.IPv6 = helpers.DefaultPointer(u.IPv6, false)

	const defaultVerbosityLevel = 1
	u.VerbosityLevel = helpers.DefaultPointer(u.VerbosityLevel, defaultVerbosityLevel)

	const defaultVerbosityDetailsLevel = 0
	u.VerbosityDetailsLevel = helpers.DefaultPointer(u.VerbosityDetailsLevel, defaultVerbosityDetailsLevel)

	const defaultValidationLogLevel = 0
	u.ValidationLogLevel = helpers.DefaultPointer(u.ValidationLogLevel, defaultValidationLogLevel)

	if u.Allowed == nil {
		u.Allowed = []netip.Prefix{
			netip.PrefixFrom(netip.AddrFrom4([4]byte{}), 0),
			netip.PrefixFrom(netip.AddrFrom16([16]byte{}), 0),
		}
	}

	u.Username = helpers.DefaultString(u.Username, "root")
}

var (
	ErrUnboundVerbosityLevelNotValid        = errors.New("Unbound verbosity level is not valid")
	ErrUnboundVerbosityDetailsLevelNotValid = errors.New("Unbound verbosity details level is not valid")
	ErrUnboundValidationLogLevelNotValid    = errors.New("Unbound validation log level is not valid")
)

func (u Unbound) validate() (err error) {
	for _, s := range u.Providers {
		_, err := provider.Parse(s)
		if err != nil {
			return err
		}
	}

	const maxVerbosityLevel = 5
	if *u.VerbosityLevel > maxVerbosityLevel {
		return fmt.Errorf("%w: %d must be between 0 and %d",
			ErrUnboundVerbosityLevelNotValid,
			*u.VerbosityLevel,
			maxVerbosityLevel)
	}

	const maxVerbosityDetailsLevel = 4
	if *u.VerbosityDetailsLevel > maxVerbosityDetailsLevel {
		return fmt.Errorf("%w: %d must be between 0 and %d",
			ErrUnboundVerbosityDetailsLevelNotValid,
			*u.VerbosityDetailsLevel,
			maxVerbosityDetailsLevel)
	}

	const maxValidationLogLevel = 2
	if *u.ValidationLogLevel > maxValidationLogLevel {
		return fmt.Errorf("%w: %d must be between 0 and %d",
			ErrUnboundValidationLogLevelNotValid,
			*u.ValidationLogLevel, maxValidationLogLevel)
	}

	return nil
}

func (u Unbound) copy() (copied Unbound) {
	return Unbound{
		Providers:             helpers.CopySlice(u.Providers),
		Caching:               helpers.CopyPointer(u.Caching),
		IPv6:                  helpers.CopyPointer(u.IPv6),
		VerbosityLevel:        helpers.CopyPointer(u.VerbosityLevel),
		VerbosityDetailsLevel: helpers.CopyPointer(u.VerbosityDetailsLevel),
		ValidationLogLevel:    helpers.CopyPointer(u.ValidationLogLevel),
		Username:              u.Username,
		Allowed:               helpers.CopySlice(u.Allowed),
	}
}

func (u *Unbound) mergeWith(other Unbound) {
	u.Providers = helpers.MergeSlices(u.Providers, other.Providers)
	u.Caching = helpers.MergeWithPointer(u.Caching, other.Caching)
	u.IPv6 = helpers.MergeWithPointer(u.IPv6, other.IPv6)
	u.VerbosityLevel = helpers.MergeWithPointer(u.VerbosityLevel, other.VerbosityLevel)
	u.VerbosityDetailsLevel = helpers.MergeWithPointer(u.VerbosityDetailsLevel, other.VerbosityDetailsLevel)
	u.ValidationLogLevel = helpers.MergeWithPointer(u.ValidationLogLevel, other.ValidationLogLevel)
	u.Username = helpers.MergeWithString(u.Username, other.Username)
	u.Allowed = helpers.MergeSlices(u.Allowed, other.Allowed)
}

func (u *Unbound) overrideWith(other Unbound) {
	u.Providers = helpers.OverrideWithSlice(u.Providers, other.Providers)
	u.Caching = helpers.OverrideWithPointer(u.Caching, other.Caching)
	u.IPv6 = helpers.OverrideWithPointer(u.IPv6, other.IPv6)
	u.VerbosityLevel = helpers.OverrideWithPointer(u.VerbosityLevel, other.VerbosityLevel)
	u.VerbosityDetailsLevel = helpers.OverrideWithPointer(u.VerbosityDetailsLevel, other.VerbosityDetailsLevel)
	u.ValidationLogLevel = helpers.OverrideWithPointer(u.ValidationLogLevel, other.ValidationLogLevel)
	u.Username = helpers.OverrideWithString(u.Username, other.Username)
	u.Allowed = helpers.OverrideWithSlice(u.Allowed, other.Allowed)
}

func (u Unbound) ToUnboundFormat() (settings unbound.Settings, err error) {
	providers := make([]provider.Provider, len(u.Providers))
	for i := range providers {
		providers[i], err = provider.Parse(u.Providers[i])
		if err != nil {
			return settings, err
		}
	}

	const port = 53

	return unbound.Settings{
		ListeningPort:         port,
		IPv4:                  true,
		Providers:             providers,
		Caching:               *u.Caching,
		IPv6:                  *u.IPv6,
		VerbosityLevel:        *u.VerbosityLevel,
		VerbosityDetailsLevel: *u.VerbosityDetailsLevel,
		ValidationLogLevel:    *u.ValidationLogLevel,
		AccessControl: unbound.AccessControlSettings{
			Allowed: netipPrefixesToNetaddrIPPrefixes(u.Allowed),
		},
		Username: u.Username,
	}, nil
}

var (
	ErrConvertingNetip = errors.New("converting net.IP to netip.Addr failed")
)

func (u Unbound) GetFirstPlaintextIPv4() (ipv4 netip.Addr, err error) {
	s := u.Providers[0]
	provider, err := provider.Parse(s)
	if err != nil {
		return ipv4, err
	}

	ip := provider.DNS().IPv4[0]
	ipv4, ok := netip.AddrFromSlice(ip)
	if !ok {
		return ipv4, fmt.Errorf("%w: for ip %s (%#v)",
			ErrConvertingNetip, ip, ip)
	}
	return ipv4.Unmap(), nil
}

func (u Unbound) String() string {
	return u.toLinesNode().String()
}

func (u Unbound) toLinesNode() (node *gotree.Node) {
	node = gotree.New("Unbound settings:")

	authServers := node.Appendf("Authoritative servers:")
	for _, provider := range u.Providers {
		authServers.Appendf(provider)
	}

	node.Appendf("Caching: %s", helpers.BoolPtrToYesNo(u.Caching))
	node.Appendf("IPv6: %s", helpers.BoolPtrToYesNo(u.IPv6))
	node.Appendf("Verbosity level: %d", *u.VerbosityLevel)
	node.Appendf("Verbosity details level: %d", *u.VerbosityDetailsLevel)
	node.Appendf("Validation log level: %d", *u.ValidationLogLevel)
	node.Appendf("System user: %s", u.Username)

	allowedNetworks := node.Appendf("Allowed networks:")
	for _, network := range u.Allowed {
		allowedNetworks.Appendf(network.String())
	}

	return node
}
