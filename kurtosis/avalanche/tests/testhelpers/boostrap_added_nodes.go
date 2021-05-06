// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package testhelpers

import (
	"fmt"

	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

func BootstrapAddedNodes(
	network networks.Network,
	definedNetwork *networkbuilder.Network,
	avalancheImage string,
	numNodes int) {

	logrus.Infof("Adding %d additional nodes and waiting for them to bootstrap...", numNodes)
	avalancheNetwork := networksavalanche.Cast(network)

	for i := 1; i <= numNodes; i++ {
		node := networkbuilder.NewNode(fmt.Sprintf("newNode-%d", i)).
			Image(avalancheImage).
			IsStaking(true)
		definedNetwork.AddNode(node)

		_, err := avalancheNetwork.CreateNode(definedNetwork, node)
		if err != nil {
			panic(stacktrace.Propagate(err, "Unable to create node %s", node.ID))
		}
		logrus.Infof("%s finished bootstrapping.", node.ID)
	}
}
