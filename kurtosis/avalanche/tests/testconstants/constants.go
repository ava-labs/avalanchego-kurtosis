// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package testconstants

import (
	"time"

	"github.com/ava-labs/avalanchego/utils/units"
)

const (
	GenesisUsername          = "genesis"
	GenesisPassword          = "MyNameIs!Jeff"
	StakerUsername           = "staker"
	StakerPassword           = "test34test!23"
	DelegatorUsername        = "delegator"
	DelegatorPassword        = "test34test!23"
	TotalAmount              = 10 * units.KiloAvax
	SeedAmount               = 5 * units.KiloAvax
	StakeAmount              = 3 * units.KiloAvax
	TxFee                    = 1 * units.Avax
	ValidatorNodeName string = "validator-node"
	DelegatorNodeName string = "delegator-node"
	TestTimeout              = 10 * time.Minute
	TestSetupTimeout         = 5 * time.Minute
)
