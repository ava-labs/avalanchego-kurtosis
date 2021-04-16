package constants

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
)

// AvalancheLogLevel specifies the log level for an Avalanche client
type AvalancheLogLevel string

const (
	// Log levels
	VERBOSE AvalancheLogLevel = "verbo"
	DEBUG   AvalancheLogLevel = "debug"
	INFO    AvalancheLogLevel = "info"

	NetworkID            uint32 = 12345
	StakingTLSCertFileID        = "staking-tls-cert"
	StakingTLSKeyFileID         = "staking-tls-key"

	TimeoutDuration = 30 * time.Second

	DefaultPassword = "This1sSuper!S4f3!..."
)

var (
	// XChainID ...
	XChainID ids.ID
	// PlatformChainID ...
	PlatformChainID ids.ID
	// CChainID ...
	CChainID ids.ID
	// AvaxAssetID ...
	AvaxAssetID ids.ID
)

func init() {
	XChainID, _ = ids.FromString("2eNy1mUFdmaxXNj1eQHUe7Np4gju9sJsEtWQ4MX3ToiNKuADed")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("WKNkfmNxgqpKPe9Q12UCoTuGYXX5JbQn2tf2WTpNTJeQrezqa")
	AvaxAssetID, _ = ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
}
