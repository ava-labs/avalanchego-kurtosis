// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package scenarios

import (
	"fmt"

	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
)

// NewBootStrappingNodeNetwork creates a new scenario with five nodes network
func NewBootStrappingNodeNetwork(avalancheImage string) *networkbuilder.Network {
	newNetwork := networkbuilder.New().
		Image(avalancheImage).
		SnowSize(3, 3)

	i := 1
	for _, staker := range constants.DefaultLocalNetGenesisConfig.Stakers {
		newNetwork.AddNode((networkbuilder.NewNode(fmt.Sprintf("bootstrapNode-%d", i)).
			Image(avalancheImage).
			IsStaking(true).
			BootstrapNode(true).
			BootstrapNodeID(i).
			ConnectedBTNodeIDs(newNetwork.GetConnectedBTNodeIDs()).
			PrivateKey(staker.PrivateKey)).
			TLSCert(staker.TLSCert),
		)
		newNetwork.ConnectedBTNodeIDs(staker.NodeID)
		i++
	}

	return newNetwork.HasBootstrapNodes(true)
}
