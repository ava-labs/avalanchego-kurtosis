// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package networkbuilder

import (
	"fmt"
	"strings"
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego/utils/units"
)

// Network defines the Network structure of the Nodes in the Topology
type Network struct {
	Nodes              map[string]*Node
	snowQuorumSize     int
	snowSampleSize     int
	image              string
	txFee              uint64
	hasBootstrapNodes  bool
	connectedBTNodeIDs []string
	connectedBTNodeIPs []string
}

// New creates the Network builder
func New() *Network {
	return &Network{
		Nodes: map[string]*Node{},
		// assumes some defaults
		txFee: 1 * units.Avax,
		image: "avaplatform/avalanchego:dev",
	}
}

func (n *Network) TxFee(txFee uint64) *Network {
	n.txFee = txFee
	return n
}

func (n *Network) SnowSize(snowSampleSize int, snowQuorumSize int) *Network {
	n.snowQuorumSize = snowQuorumSize
	n.snowSampleSize = snowSampleSize
	return n
}

func (n *Network) Image(image string) *Network {
	n.image = image
	return n
}

func (n *Network) AddNode(node *Node) *Network {
	if _, ok := n.Nodes[node.ID]; ok {
		panic("Node already exist")
	}

	n.Nodes[node.ID] = node.ConnectedBTNodeIDs(n.GetConnectedBTNodeIDs())
	return n
}

func (n *Network) GetTxFee() uint64 {
	return n.txFee
}

func (n *Network) GetSnowQuorumSize() int {
	return n.snowQuorumSize
}

func (n *Network) GetSnowSampleSize() int {
	return n.snowSampleSize
}

func (n *Network) HasBootstrapNodes(b bool) *Network {
	n.hasBootstrapNodes = true
	return n
}

func (n *Network) ConnectedBTNodeIDs(btNodeID string) *Network {
	n.connectedBTNodeIDs = append(n.connectedBTNodeIDs, btNodeID)
	return n
}

func (n *Network) GetConnectedBTNodeIDs() string {
	return strings.Join(n.connectedBTNodeIDs, ",")
}

func (n *Network) RemoveNode(node *Node) *Network {
	delete(n.Nodes, node.ID)
	return n
}

func (n *Network) GetNumBootstrapNodes() int {
	return len(n.connectedBTNodeIDs)
}

type Node struct {
	ID                    string
	varyCerts             bool
	serviceLogLevel       constants.AvalancheLogLevel
	imageName             string
	snowQuorumSize        int
	snowSampleSize        int
	networkInitialTimeout time.Duration
	isStaking             bool
	privateKey            string
	tlsCert               string
	isBootstrapNode       bool
	connectedBTNodes      string
	bootstrapNodeID       int
	connectedBTNodeIPs    string
	boostrapAttempts      int
}

func NewNode(nodeID string) *Node {
	return &Node{
		ID:                    nodeID,
		varyCerts:             true,
		serviceLogLevel:       constants.DEBUG,
		imageName:             "avaplatform/avalanchego:v1.2.3",
		snowQuorumSize:        1,
		snowSampleSize:        1,
		networkInitialTimeout: 2 * time.Second,
		boostrapAttempts:      10,
	}
}

func (node *Node) Image(imageName string) *Node {
	node.imageName = imageName
	return node
}

func (node *Node) GetImage() string {
	return node.imageName
}

func (node *Node) IsStaking(b bool) *Node {
	node.isStaking = b
	return node
}

func (node *Node) GetStaking() bool {
	return node.isStaking
}

func (node *Node) PrivateKey(key string) *Node {
	node.privateKey = key
	return node
}

func (node *Node) GetPrivateKey() string {
	return node.privateKey
}

func (node *Node) TLSCert(cert string) *Node {
	node.tlsCert = cert
	return node
}

func (node *Node) GetTLSCert() string {
	return node.tlsCert
}

func (node *Node) HasCerts() bool {
	return len(node.tlsCert) > 0 && len(node.privateKey) > 0
}

func (node *Node) BootstrapNode(bootstrap bool) *Node {
	node.isBootstrapNode = bootstrap
	return node
}

func (node *Node) IsBootstrapNode() bool {
	return node.isBootstrapNode
}

func (node *Node) ConnectedBTNodeIDs(btNodes string) *Node {
	node.connectedBTNodes = btNodes
	return node
}

func (node *Node) GetConnectedBTNodeIDs() string {
	return node.connectedBTNodes
}

func (node *Node) BootstrapNodeID(i int) *Node {
	node.bootstrapNodeID = i
	return node
}

func (node *Node) GetBootstrapNodeID() int {
	return node.bootstrapNodeID
}

func (node *Node) String() string {
	return fmt.Sprintf("NodeID: %s, HasCerts: %v, serviceLogLevel: %s, imageName: %s, snowQuorumSize: %d,"+
		"snowSampleSize: %d, networkInitialTimeout: %v, isStaking: %v, isBootstrapNode: %v, connectedBTNodes: %s"+
		"bootstrapNodeID: %d",
		node.ID, node.varyCerts, node.serviceLogLevel, node.imageName, node.snowQuorumSize, node.snowSampleSize,
		node.networkInitialTimeout, node.isStaking, node.isBootstrapNode, node.connectedBTNodes, node.bootstrapNodeID,
	)
}

func (node *Node) BootstrapAttempts(i int) *Node {
	node.boostrapAttempts = i
	return node
}

func (node *Node) GetBootstrapAttempts() int {
	return node.boostrapAttempts
}

func (node *Node) GetStakingPort() int {
	return 9651
}

func (node *Node) GetHTTPPort() int {
	return 9650
}
