// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chainhelper

import (
	"time"

	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/ava-labs/avalanchego/ids"
)

// This helper automates some the most used functions in the CChain
type CChainHelper struct {
}

// AwaitTransactionAcceptance waits for the [txID] to be accepted within [timeout]
func (c *CChainHelper) AwaitTransactionAcceptance(client *avalanchegoclient.Client, txID ids.ID, timeout time.Duration) error {

	// TODO replace when getTxStatus is added to the C Chain API
	time.Sleep(timeout / 5)
	return nil
}

// CheckBalance validates the [address] balance is equal to [amount]
func (c *CChainHelper) CheckBalance(client *avalanchegoclient.Client, address string, assetID string, expectedAmount uint64) error {
	panic("TODO")
}

// CChain is a helper to chain request to the correct VM
func CChain() *CChainHelper {
	return &CChainHelper{}
}
