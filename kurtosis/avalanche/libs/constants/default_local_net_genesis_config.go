// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package constants

// DefaultLocalNetGenesisConfig contains the private keys and node IDs that come from avalanchego for the 5 bootstrapper nodes.
// When using avalanchego with the 'local' testnet option, the P-chain comes preloaded with five bootstrapper nodes whose node
// IDs are hardcoded in avalanchego source. Node IDs are determined based off the TLS keys of the nodes, so to ensure that
// we can launch nodes with the same node ID (to validate, else we wouldn't be able to validate at all), the avalanchego
// source code also provides the private keys for these nodes.
var DefaultLocalNetGenesisConfig = NetworkGenesisConfig{
	Stakers: defaultStakers,
	// hardcoded in avalanchego in "genesis/config.go". needed to distribute genesis funds in tests
	FundedAddresses: FundedAddress{
		"X-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u",
		/*
			 	It's okay to have privateKey here because its a hardcoded value available in the avalanchego codebase.
				It is necessary to have this privateKey in order to transfer funds to test accounts in the test net.
				This privateKey only applies to local test nets, it has nothing to do with the public test net or main net.
		*/
		"PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN",
	},
}

/*
Snow consensus requires at least $snow_consensus stakers for for liveness. But, you can't register new stakers... without
meeting that threshold. Thus, some stakers are hardcoded in the genesis.
https://github.com/ava-labs/avalanchego/blob/master/genesis/config.go#L662

These IDs are those stakers for the default local network config
*/
var defaultStakers = []StakerIdentity{
	staker1,
	staker2,
	staker3,
	staker4,
	staker5,
}

var staker1 = StakerIdentity{
	Staker1NodeID,
	Staker1PrivateKey,
	Staker1Cert,
	"",
}

var staker2 = StakerIdentity{
	Staker2NodeID,
	Staker2PrivateKey,
	Staker2Cert,
	Staker1NodeID,
}

var staker3 = StakerIdentity{
	Staker3NodeID,
	Staker3PrivateKey,
	Staker3Cert,
	Staker1NodeID + "," + Staker2NodeID,
}

var staker4 = StakerIdentity{
	Staker4NodeID,
	Staker4PrivateKey,
	Staker4Cert,
	Staker1NodeID + "," + Staker2NodeID + "," + Staker3NodeID,
}

var staker5 = StakerIdentity{
	Staker5NodeID,
	Staker5PrivateKey,
	Staker5Cert,
	Staker1NodeID + "," + Staker2NodeID + "," + Staker3NodeID + "," + Staker4NodeID,
}
