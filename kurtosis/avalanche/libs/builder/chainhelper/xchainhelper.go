// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chainhelper

import (
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// This helper automates some the most used functions in the XChain
type XChainHelper struct{}

// AwaitTransactionAcceptance waits for the [txID] to be accepted within [timeout]
func (x *XChainHelper) AwaitTransactionAcceptance(client *avalanchegoclient.Client, txID ids.ID, timeout time.Duration) error {

	for startTime := time.Now(); time.Since(startTime) < timeout; time.Sleep(time.Second) {
		status, err := client.XChainAPI().GetTxStatus(txID)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get status.")
		}
		logrus.Tracef("Status for transaction %s: %s", txID, status)
		if status == choices.Accepted {
			return nil
		}
		if status == choices.Rejected {
			return stacktrace.NewError("Transaction %s was %s", txID, status)
		}
	}
	return stacktrace.NewError("Timed out waiting for transaction %s to be accepted on the XChain.", txID)
}

// CheckBalance validates the [address] balance is equal to [amount]
func (x *XChainHelper) CheckBalance(client *avalanchegoclient.Client, address string, assetID string, expectedAmount uint64) error {
	xBalance, err := client.XChainAPI().GetBalance(address, assetID, false)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to retrieve X Chain balance.")
	}
	xActualBalance := uint64(xBalance.Balance)
	if xActualBalance != expectedAmount {
		return stacktrace.NewError("Found unexpected X Chain Balance for address: %s. Expected: %v, found: %v",
			address, expectedAmount, xActualBalance)
	}

	return nil
}

// XChain is a helper to chain request to the correct VM
func XChain() *XChainHelper {
	return &XChainHelper{}
}
