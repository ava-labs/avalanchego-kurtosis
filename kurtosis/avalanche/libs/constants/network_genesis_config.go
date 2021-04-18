// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package constants

// NetworkGenesisConfig encapsulates genesis information describing
// a network
type NetworkGenesisConfig struct {
	Stakers         []StakerIdentity
	FundedAddresses FundedAddress
}

// FundedAddress encapsulates a pre-funded address
type FundedAddress struct {
	Address    string
	PrivateKey string
}

// StakerIdentity contains a staker's identifying information
type StakerIdentity struct {
	NodeID         string
	PrivateKey     string
	TLSCert        string
	ConnectedNodes string
}
