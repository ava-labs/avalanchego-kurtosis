package topology

import (
	"fmt"
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/chainhelper"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	cjson "github.com/ava-labs/avalanchego/utils/json"
)

// Genesis is a single attribute of the Topology (only one Genesis) that holds the Genesis data
type Genesis struct {
	id       string
	client   *avalanchegoclient.Client
	Address  string
	userPass api.UserPass
	context  *testsuite.TestContext
}

func newGenesis(id string, userName string, password string, client *avalanchegoclient.Client, context *testsuite.TestContext) *Genesis {
	return &Genesis{
		id: id,
		userPass: api.UserPass{
			Username: userName,
			Password: password,
		},
		client:  client,
		context: context,
	}
}

// ImportGenesisFunds fetches the default funded funds and imports them
func (g *Genesis) ImportGenesisFunds() error {
	var err error

	keystore := g.client.KeystoreAPI()
	if _, err = keystore.CreateUser(g.userPass); err != nil {
		return stacktrace.Propagate(err, "Failed to take create genesis user account.")
	}

	g.Address, err = g.client.XChainAPI().ImportKey(
		g.userPass,
		constants.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to take control of genesis account.")
	}
	logrus.Infof("Genesis Address: %s.", g.Address)
	return nil
}

// FundXChainAddresses funds the genesis funds into an address on the XChain
func (g *Genesis) FundXChainAddresses(addresses []string, amount uint64) *Genesis {
	for _, address := range addresses {
		txID, err := g.client.XChainAPI().Send(
			g.userPass,
			nil,    // from addrs
			"",     // change addr
			amount, // deducted = ( amount + txFee)
			"AVAX",
			address,
			"",
		)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to fund addresses with genesis funds."))
			return g
		}

		// wait for the tx to go through
		err = chainhelper.XChain().AwaitTransactionAcceptance(g.client, txID, constants.TimeoutDuration)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Timed out waiting for transaction to be accepted on the XChain"))
			return g
		}

		// verify the balance
		err = chainhelper.XChain().CheckBalance(g.client, address, "AVAX", amount)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to validate fund on the XChain"))
			return g
		}

		logrus.Infof("Funded X Chain Address: %s with %d.", address, amount)
	}

	return g
}

func (g *Genesis) MultipleFundXChainAddresses(addresses []string, amount uint64, times int) *Genesis {
	txIDs := make([]ids.ID, len(addresses))
	for i, address := range addresses {
		// create the multiple outputs
		sendOutputs := make([]avm.SendOutput, 0, times)
		for j := 0; j < times; j++ {
			sendOutputs = append(sendOutputs, avm.SendOutput{
				AssetID: "AVAX",
				Amount:  cjson.Uint64(amount),
				To:      address,
			})
		}

		// send it
		txID, err := g.client.XChainAPI().SendMultiple(g.userPass, nil, "", sendOutputs, "")
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to send transaction with %d outputs", len(sendOutputs)))
			return g
		}

		logrus.Infof("Sent transaction %s with %d outputs", txID, len(sendOutputs))

		// wait for the transactions to be accepted
		err = chainhelper.XChain().AwaitTransactionAcceptance(g.client, txID, constants.TimeoutDuration)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to wait transaction accepted for address %s", addresses[i]))
		}
		logrus.Infof("Transaction: %v accepted for address %s with %d utxos", txID.String(), address, len(sendOutputs))

		txIDs[i] = txID
	}

	logrus.Infof("Expected %d addresses to have %d utxos", len(addresses), len(txIDs))

	var errG errgroup.Group
	startTime := time.Now()
	// check the balance
	logrus.Infof("Expected amount on each address of : %v", amount*uint64(times))
	for _, address := range addresses {
		errG.Go(func() error {
			// verify the balance
			err := chainhelper.XChain().CheckBalance(g.client, address, "AVAX", amount*uint64(times))
			if err != nil {
				return err
			}

			return nil
		})
	}
	err := errG.Wait()
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "Failed to check funds in addresses"))
	}

	logrus.Infof("Funded X Chain Addresses: %d with %d in %v seconds.", len(addresses), amount, time.Since(startTime).Seconds())
	return g
}

func (g *Genesis) MultipleFundXChainAddresses2(addresses []string, amount uint64, times int) *Genesis {
	txIDs := make([]ids.ID, times)
	// create the multiple outputs
	sendOutputs := make([]avm.SendOutput, 0, times)
	for _, address := range addresses {
		for j := 0; j < times; j++ {
			sendOutputs = append(sendOutputs, avm.SendOutput{
				AssetID: "AVAX",
				Amount:  cjson.Uint64(amount),
				To:      address,
			})
		}
	}

	// send it
	txID, err := g.client.XChainAPI().SendMultiple(g.userPass, nil, "", sendOutputs, "")
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "Failed to send transaction with %d outputs", len(sendOutputs)))
		return g
	}

	logrus.Infof("Sent 1 transaction %s with %d outputs", txID, len(sendOutputs))

	// wait for the transactions to be accepted
	err = chainhelper.XChain().AwaitTransactionAcceptance(g.client, txID, constants.TimeoutDuration)
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "Failed to wait transaction accepted"))
	}
	logrus.Infof("Transaction: %v accepted for addresses %v with %d utxos", txID, addresses, len(sendOutputs))

	logrus.Infof("Expected %d addresses to have %d utxos", len(addresses), len(txIDs))

	var errG errgroup.Group
	startTime := time.Now()
	// check the balance
	logrus.Infof("Expected amount on each address of : %v", amount*uint64(times))
	for _, address := range addresses {
		errG.Go(func() error {
			// verify the balance
			err := chainhelper.XChain().CheckBalance(g.client, address, "AVAX", amount*uint64(times))
			if err != nil {
				return err
			}

			return nil
		})
	}
	err = errG.Wait()
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "Failed to check funds in addresses"))
	}

	logrus.Infof("Funded X Chain Addresses: %d with %d in %v seconds.", len(addresses), amount, time.Since(startTime).Seconds())
	return g
}

func (g *Genesis) FundCChainAddresses(addrs []common.Address, amount uint64) {
	logrus.Infof("Using address : %v", g.Address)
	_, err := g.client.CChainAPI().ImportKey(
		g.userPass,
		constants.DefaultLocalNetGenesisConfig.FundedAddresses.PrivateKey)
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "unable to fund cchain"))
	}

	for _, addr := range addrs {
		cChainBech32 := fmt.Sprintf("C%s", g.Address[1:])
		txID, err := g.client.XChainAPI().ExportAVAX(g.userPass, nil, "", amount, cChainBech32)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to C-Chain"))
		}
		err = chainhelper.XChain().AwaitTransactionAcceptance(g.client, txID, constants.TimeoutDuration)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Timed out waiting to export AVAX to C-Chain"))

		}

		txID, err = g.client.CChainAPI().Import(g.userPass, addr.Hex(), "X")
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Failed to import AVAX to C-Chain"))
		}

		err = chainhelper.CChain().AwaitTransactionAcceptance(g.client, txID, constants.TimeoutDuration)
		if err != nil {
			g.context.Fatal(stacktrace.Propagate(err, "Timed out waiting to import AVAX to C-Chain"))

		}
	}
}

func (g *Genesis) MoveBalanceToCChain(addr common.Address, txFee uint64) {
	balance, err := g.client.XChainAPI().GetBalance(g.Address, constants.AvaxAssetID.String(), true)
	if err != nil {
		g.context.Fatal(stacktrace.Propagate(err, "Unable to fetch balance from the genesis X chain address"))
	}

	sendableBalance := (uint64(balance.Balance) - txFee) / 2
	logrus.Infof("Balance: %v, txFee: %v,Sendable : %v", uint64(balance.Balance), txFee, sendableBalance)

	g.FundCChainAddresses([]common.Address{addr}, sendableBalance)
}
