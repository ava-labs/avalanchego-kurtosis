// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package topologyhelper

import (
	"fmt"
	"strings"

	top "github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/topology"
)

func VerifyConnectedPeers(n1 []*top.Node, n2 []*top.Node) error {
	for _, node := range n1 {
		peers, err := node.GetClient().InfoAPI().Peers()
		if err != nil {
			return err
		}

		for _, peer := range peers {
			peerIP := strings.Split(peer.IP, ":")[0]
			found := false
			for _, comparingNode := range n2 {
				if peerIP == comparingNode.GetIPAddress() {
					found = true
					break
				}
			}
			if !found {
				addresses := map[string]string{}
				for _, nod := range n2 {
					addresses[nod.NodeID] = nod.GetIPAddress()
				}
				return fmt.Errorf(
					"address IP: %s, not found in the peer list: %v",
					peerIP,
					addresses,
				)
			}
		}
	}
	return nil
}
