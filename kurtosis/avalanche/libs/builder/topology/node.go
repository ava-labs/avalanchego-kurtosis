package topology

import (
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/chainhelper"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// Node defines the Node in the Topology
type Node struct {
	id        string
	UserPass  api.UserPass
	PAddress  string
	XAddress  string
	client    *avalanchegoclient.Client
	context   *testsuite.TestContext
	NodeID    string
	ipAddress string
}

func newNode(id string, userName string, password string, ipAddress string, client *avalanchegoclient.Client, context *testsuite.TestContext) *Node {
	nodeID, err := client.InfoAPI().GetNodeID()
	if err != nil {
		context.Fatal(stacktrace.Propagate(err, "Could not get node ID."))
	}

	return &Node{
		id: id,
		UserPass: api.UserPass{
			Username: userName,
			Password: password,
		},
		NodeID:    nodeID,
		client:    client,
		context:   context,
		ipAddress: ipAddress,
	}
}

// CreateAddress creates user and both XChain and PChain addresses for the Node
func (n *Node) CreateAddress() *Node {
	keystore := n.client.KeystoreAPI()
	if _, err := keystore.CreateUser(n.UserPass); err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Could not create user for node."))
	}

	xAddress, err := n.client.XChainAPI().CreateAddress(n.UserPass)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Could not create user address in the XChainAPI."))
	}
	n.XAddress = xAddress

	pAddress, err := n.client.PChainAPI().CreateAddress(n.UserPass)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Could not create user address in the PChainAPI."))
	}
	n.PAddress = pAddress
	return n
}

// GetClient returns the RPC API client to access the nodes VMS
func (n *Node) GetClient() *avalanchegoclient.Client {
	return n.client
}

// BecomeValidator is a multi step methods that does the following
// - exports AVAX from the XChain + waits for acceptance in the XChain
// - imports the amount to the PChain + waits for acceptance in the PChain
// - verifies the PChain balance + verifies the XChain balance
// - adds nodeID as a validator - waits Tx acceptance in the PChain
// - waits until the validation period begins
//
func (n *Node) BecomeValidator(genesisAmount uint64, seedAmount uint64, stakeAmount uint64, txFee uint64) *Node {
	// exports AVAX from the X Chain
	exportTxID, err := n.client.XChainAPI().ExportAVAX(
		n.UserPass,
		nil,              // from addrs
		"",               // change addr
		seedAmount+txFee, // deducted (seedAmmount + txFee(ExportAVAX) ) - 1xFee deducted from XChain + 1xFee to be deducted from PChain Tx
		n.PAddress,
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to pchainAddress %s", n.PAddress))
		return n
	}

	// waits Tx acceptance in the XChain
	err = chainhelper.XChain().AwaitTransactionAcceptance(n.client, exportTxID, 120*time.Second)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX from XChain Address %s", n.XAddress))
		return n
	}

	// imports the amount to the P Chain
	importTxID, err := n.client.PChainAPI().ImportAVAX( // receivedAmount = (sent - txFee)
		n.UserPass,
		nil, // from addrs
		"",  // change addr
		n.PAddress,
		constants.XChainID.String(),
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed import AVAX to PChain Address %s", n.PAddress))
		return n
	}

	// waits Tx acceptance in the PChain
	err = chainhelper.PChain().AwaitTransactionAcceptance(n.client, importTxID, constants.TimeoutDuration)
	if err != nil {
		n.context.Fatal(err)
		return n
	}

	// verify the PChain balance of seedAmount on the PChain (which should have been at 0)
	err = chainhelper.PChain().CheckBalance(n.client, n.PAddress, seedAmount) // balance = seedAmount = transferred + txFee
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "expected balance of seedAmount the stakeAmount was moved to XChain"))
		return n
	}

	// verify the XChain balance of (seedAmount - stakeAmount - 2*txFee) the stake was moved to PChain
	err = chainhelper.XChain().CheckBalance(n.client, n.XAddress, "AVAX", genesisAmount-seedAmount-2*txFee)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "expected balance of (seedAmount - stakeAmount - 2*txFee) the stake was moved to XChain"))
		return n
	}

	// add nodeID as a validator
	stakingStartTime := time.Now().Add(20 * time.Second)
	startTime := uint64(stakingStartTime.Unix())
	endTime := uint64(stakingStartTime.Add(72 * time.Hour).Unix())
	addStakerTxID, err := n.client.PChainAPI().AddValidator(
		n.UserPass,
		nil,
		"",
		n.PAddress,
		n.NodeID,
		stakeAmount,
		startTime,
		endTime,
		float32(2),
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to add validator to primary network %s", n.id))
		return n
	}

	// waits Tx acceptance in the PChain
	err = chainhelper.PChain().AwaitTransactionAcceptance(n.client, addStakerTxID, constants.TimeoutDuration)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "transaction not accepted"))
		return n
	}

	// waits until the validation period begins
	time.Sleep(time.Until(stakingStartTime) + 3*time.Second)

	// verifies if the node is a current validator
	currentStakers, err := n.client.PChainAPI().GetCurrentValidators(ids.Empty)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Could not get current stakers."))
		return n
	}

	found := false
	for _, stakerIntf := range currentStakers {
		staker := stakerIntf.(map[string]interface{})
		if staker["nodeID"] == n.NodeID {
			found = true
			break
		}
	}

	if !found {
		n.context.Fatal(stacktrace.NewError("Node: %s not found in the stakers %v", n.NodeID, currentStakers))
		return n
	}

	// verifies the balance of the staker in the PChain - should be the seedAmmount - stakedAmount
	err = chainhelper.PChain().CheckBalance(n.client, n.PAddress, seedAmount-stakeAmount)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Error checking the PChain balance."))
	}

	logrus.Infof("Verified the staker was added to current validators and has the expected P Chain balance.")

	return n
}

// BecomeDelegator is a multi step methods that does the following
// - exports AVAX from the XChain + waits for acceptance in the XChain
// - imports the amount to the PChain + waits for acceptance in the PChain
// - verifies the PChain balance + verifies the XChain balance
// - adds nodeID as a delegator - waits Tx acceptance in the PChain
// - waits until the validation period begins
//
func (n *Node) BecomeDelegator(genesisAmount uint64, seedAmount uint64, delegatorAmount uint64, txFee uint64, stakerNodeID string) *Node {

	// exports AVAX from the X Chain
	exportTxID, err := n.client.XChainAPI().ExportAVAX(
		n.UserPass,
		nil, // from addrs
		"",  // change addr
		seedAmount+txFee,
		n.PAddress,
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to export AVAX to pchainAddress %s", n.PAddress))
		return n
	}

	// waits Tx acceptance in the XChain
	err = chainhelper.XChain().AwaitTransactionAcceptance(n.client, exportTxID, constants.TimeoutDuration)
	if err != nil {
		n.context.Fatal(err)
		return n
	}

	// imports the amount to the P Chain
	importTxID, err := n.client.PChainAPI().ImportAVAX(
		n.UserPass,
		nil, // from addrs
		"",  // change addr
		n.PAddress,
		constants.XChainID.String(),
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed import AVAX to pchainAddress %s", n.PAddress))
		return n
	}

	// waits Tx acceptance in the PChain
	err = chainhelper.PChain().AwaitTransactionAcceptance(n.client, importTxID, constants.TimeoutDuration)
	if err != nil {
		n.context.Fatal(err)
		return n
	}

	// verify the PChain balance (seedAmount+txFee-txFee)
	err = chainhelper.PChain().CheckBalance(n.client, n.PAddress, seedAmount)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "expected balance of seedAmount exists in the PChain"))
		return n
	}

	// verify the XChain balance of genesisAmount - seedAmount - txFee - txFee (import PChain)
	err = chainhelper.XChain().CheckBalance(n.client, n.XAddress, "AVAX", genesisAmount-seedAmount-2*txFee)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "expected balance XChain balance of genesisAmount-seedAmount-txFee"))
		return n
	}

	delegatorStartTime := time.Now().Add(20 * time.Second)
	startTime := uint64(delegatorStartTime.Unix())
	endTime := uint64(delegatorStartTime.Add(36 * time.Hour).Unix())
	addDelegatorTxID, err := n.client.PChainAPI().AddDelegator(
		n.UserPass,
		nil, // from addrs
		"",  // change addr
		n.PAddress,
		stakerNodeID,
		delegatorAmount,
		startTime,
		endTime,
	)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to add delegator %s", n.PAddress))
		return n
	}

	err = chainhelper.PChain().AwaitTransactionAcceptance(n.client, addDelegatorTxID, constants.TimeoutDuration)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Failed to accept AddDelegator tx: %s", addDelegatorTxID))
		return n
	}

	// Sleep until delegator starts validating
	time.Sleep(time.Until(delegatorStartTime) + 3*time.Second)

	expectedDelegatorBalance := seedAmount - delegatorAmount
	err = chainhelper.PChain().CheckBalance(n.client, n.PAddress, expectedDelegatorBalance)
	if err != nil {
		n.context.Fatal(stacktrace.Propagate(err, "Unexpected P Chain Balance after adding a new delegator to the network."))
		return n
	}
	logrus.Infof("Added delegator to subnet and verified the expected P Chain balance.")

	return n
}

func (n *Node) GetIPAddress() string {
	return n.ipAddress
}
