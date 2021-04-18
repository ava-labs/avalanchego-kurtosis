// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanchegoclient

import (
	"context"
	"math/big"
	"sync"

	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ethereum/go-ethereum/common"
)

// ConcurrentEthClient is a concurrency-safe implementation
// of ethclient.Client that allows for multiple concurrent
// requests to be made to a single *services.Client.
type ConcurrentEthClient struct {
	client *ethclient.Client
	lock   sync.Mutex
}

// NewConcurrentEthClient ...
func NewConcurrentEthClient(client *ethclient.Client) *ConcurrentEthClient {
	return &ConcurrentEthClient{
		client: client,
	}
}

// Close terminates the client's connection.
func (c *ConcurrentEthClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.client.Close()
}

// SendTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (c *ConcurrentEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.client.SendTransaction(ctx, tx)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (c *ConcurrentEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.client.TransactionReceipt(ctx, txHash)
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (c *ConcurrentEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.client.BalanceAt(ctx, account, blockNumber)
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (c *ConcurrentEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.client.BlockByNumber(ctx, number)
}

// BlockNumber returns the most recent block number
func (c *ConcurrentEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.client.BlockNumber(ctx)
}
