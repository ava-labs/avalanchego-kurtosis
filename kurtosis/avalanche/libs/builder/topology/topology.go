// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package topology

import (
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// Topology defines how the nodes behave/capabilities in the network
type Topology struct {
	network *networksavalanche.AvalancheNetwork
	genesis *Genesis
	nodes   map[string]*Node
}

// New creates a new instance of the Topology
func New(network networks.Network) *Topology {
	avalancheNetwork := networksavalanche.Cast(network)
	return &Topology{
		network: avalancheNetwork,
		nodes:   map[string]*Node{},
	}
}

// AddNode adds a new now with both PChain and XChain address
func (s *Topology) AddNode(id string, username string, password string) *Topology {
	client, err := s.network.GetNodeClient(id)
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to fetch the Avalanche client"))
		return s
	}

	ipAddress, err := s.network.GetIPAddress(id)
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to fetch the Avalanche node IP address"))
		return s
	}

	newNode := newNode(id, username, password, ipAddress, client).CreateAddress()
	nodeID, err := client.InfoAPI().GetNodeID()
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to fetch the InfoAPI Node ID"))
		return s
	}

	logrus.Infof("New node in the Topology - Node: %s NodeID: %s", id, nodeID)
	s.nodes[id] = newNode
	return s
}

// AddGenesis creates the Genesis property in the Topology
func (s *Topology) AddGenesis(nodeID string, username string, password string) *Topology {
	client, err := s.network.GetNodeClient(nodeID)
	if err != nil {
		panic(stacktrace.Propagate(err, "Unable to fetch the genesis Avalanche client"))
		return s
	}

	s.genesis = newGenesis(nodeID, username, password, client)
	err = s.genesis.ImportGenesisFunds()
	if err != nil {
		panic(stacktrace.Propagate(err, "Could not get delegator node ID."))
	}

	return s
}

// Genesis returns the Topology Genesis
func (s *Topology) Genesis() *Genesis {
	return s.genesis
}

// Node returns a Node given the [nodeID]
func (s *Topology) Node(nodeID string) *Node {
	return s.nodes[nodeID]
}

func (s *Topology) GetAllNodes() []*Node {
	var allNodes []*Node
	for _, node := range s.nodes {
		allNodes = append(allNodes, node)
	}
	return allNodes
}

func (s *Topology) RemoveNode(id string) *Topology {
	delete(s.nodes, id)
	return s
}

func (s *Topology) LoadDefinedNetwork(definedNetwork *networkbuilder.Network) *Topology {
	for _, node := range definedNetwork.Nodes {
		s.AddNode(node.ID, node.ID, constants.DefaultPassword)
	}
	return s
}
