// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package tests

import (
	"time"

	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/chainhelper"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/scenarios"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/tests/testconstants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/tests/testhelpers"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/testsuiteavalanche/runner"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"

	top "github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/topology"
)

func Workflow(avalancheImage string) *runner.AvalancheTestRunner {

	txFee := testconstants.TxFee
	totalAmount := testconstants.TotalAmount
	seedAmount := testconstants.SeedAmount
	stakeAmount := testconstants.StakeAmount
	validatorNodeName := testconstants.ValidatorNodeName
	stakerUsername := testconstants.StakerUsername
	stakerPassword := testconstants.StakerPassword
	delegatorNodeName := testconstants.DelegatorNodeName
	delegatorUsername := testconstants.DelegatorUsername
	delegatorPassword := testconstants.DelegatorPassword
	genesisUsername := testconstants.GenesisUsername
	genesisPassword := testconstants.GenesisPassword

	// create the nodes
	stakerNode := networkbuilder.NewNode(testconstants.ValidatorNodeName).
		Image(avalancheImage).
		IsStaking(true)

	delegatorNode := networkbuilder.NewNode(testconstants.DelegatorNodeName).
		Image(avalancheImage).
		IsStaking(true)

	definedNetwork := scenarios.NewBootStrappingNodeNetwork(avalancheImage).
		TxFee(txFee).
		AddNode(stakerNode).
		AddNode(delegatorNode)

	// the actual test
	test := func(network networks.Network, context testsuite.TestContext) {

		// builds the topology of the test
		topology := top.New(network, &context)
		topology.
			AddNode(validatorNodeName, stakerUsername, stakerPassword).
			AddNode(delegatorNodeName, delegatorUsername, delegatorPassword).
			AddGenesis(validatorNodeName, genesisUsername, genesisPassword)

		// creates a genesis and funds the X addresses of the nodes
		topology.Genesis().
			FundXChainAddresses([]string{
				topology.Node(validatorNodeName).XAddress,
				topology.Node(delegatorNodeName).XAddress,
			},
				totalAmount,
			)

		// sets the nodes to validators and delegators
		// validatorNodeName - will have available after this op :
		// XChain - 10k - 5k - 2*txFee - 4998000000000
		// PChain - 5k - 3k = 2k
		topology.Node(validatorNodeName).BecomeValidator(totalAmount, seedAmount, stakeAmount, txFee)

		// delegatorNodeName - will have available after this op :
		// XChain - 10k - 5k - 2*txFee = 4998000000000
		// PChain - 5k - 3k = 2k
		topology.Node(delegatorNodeName).BecomeDelegator(totalAmount, seedAmount, stakeAmount, txFee, topology.Node(validatorNodeName).NodeID)

		// after setup we want to test moving amounts from P to X Chain and back
		stakerNode := topology.Node(validatorNodeName)

		// Lets move whats in the PChain back to the XChain = 2k - txFee (to be burned)
		exportTxID, err := stakerNode.GetClient().PChainAPI().ExportAVAX(
			stakerNode.UserPass,
			[]string{},
			"", // change addr
			stakerNode.XAddress,
			seedAmount-stakeAmount-txFee,
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", stakerNode.XAddress))
		}

		err = chainhelper.PChain().AwaitTransactionAcceptance(stakerNode.GetClient(), exportTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID))
		}

		importTxID, err := stakerNode.GetClient().XChainAPI().ImportAVAX(
			stakerNode.UserPass,
			stakerNode.XAddress,
			constants.PlatformChainID.String())
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to import AVAX to xChainAddress %s", stakerNode.XAddress))
		}

		err = chainhelper.XChain().AwaitTransactionAcceptance(stakerNode.GetClient(), importTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to wait for acceptance of transaction on XChain."))
		}

		err = chainhelper.PChain().CheckBalance(stakerNode.GetClient(), stakerNode.PAddress, 0)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer."))
		}

		// Now we should have
		// XChain: 10k - 5k - 2*txFee (1st op) = 4998000000000 + 3k - 2*txFee (2nd export)
		err = chainhelper.XChain().CheckBalance(stakerNode.GetClient(), stakerNode.XAddress, "AVAX",
			totalAmount-seedAmount-2*txFee+seedAmount-stakeAmount-2*txFee)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer."))
		}
		logrus.Infof("Transferred leftover staker funds back to X Chain and verified X and P balances.")

		delegatorNode := topology.Node("delegator-node")

		exportTxID, err = delegatorNode.GetClient().PChainAPI().ExportAVAX(
			delegatorNode.UserPass,
			[]string{},
			"", // change addr
			delegatorNode.XAddress,
			seedAmount-stakeAmount-txFee,
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", delegatorNode.XAddress))
		}

		err = chainhelper.PChain().AwaitTransactionAcceptance(delegatorNode.GetClient(), exportTxID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to accept ExportTx: %s", exportTxID))
		}

		txID, err := delegatorNode.GetClient().XChainAPI().ImportAVAX(
			delegatorNode.UserPass,
			delegatorNode.XAddress,
			constants.PlatformChainID.String(),
		)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to xChainAddress %s", delegatorNode.XAddress))
		}

		err = chainhelper.XChain().AwaitTransactionAcceptance(delegatorNode.GetClient(), txID, 30*time.Second)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Failed to Accept ImportTx: %s", importTxID))
		}

		err = chainhelper.PChain().CheckBalance(delegatorNode.GetClient(), delegatorNode.PAddress, 0)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected P Chain Balance after P -> X Transfer."))
		}

		err = chainhelper.XChain().CheckBalance(delegatorNode.GetClient(), delegatorNode.XAddress, "AVAX",
			totalAmount-seedAmount-2*txFee+seedAmount-stakeAmount-2*txFee)
		if err != nil {
			context.Fatal(stacktrace.Propagate(err, "Unexpected X Chain Balance after P -> X Transfer."))
		}

		logrus.Infof("Transferred leftover delegator funds back to X Chain and verified X and P balances.")

		testhelpers.BootstrapAddedNodes(network, context, definedNetwork, avalancheImage, 2)
	}

	return runner.NewGenericAvalancheTestRunner(definedNetwork, test, testconstants.TestTimeout, testconstants.TestSetupTimeout)
}
