// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package runner

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type AvalancheTestRunner struct {
	nodeImage      string
	definedNetwork *networkbuilder.Network
	runnableTest   func(network networks.Network) error
	testTimeout    time.Duration
	setupTimeout   time.Duration
}

func NewGenericAvalancheTestRunner(definedNetwork *networkbuilder.Network, test func(network networks.Network) error, testTimeout time.Duration, setupTimeout time.Duration) *AvalancheTestRunner {
	return &AvalancheTestRunner{
		definedNetwork: definedNetwork,
		runnableTest:   test,
		testTimeout:    testTimeout,
		setupTimeout:   setupTimeout,
	}
}

func (runner *AvalancheTestRunner) Configure(builder *testsuite.TestConfigurationBuilder) {
	setupTimeoutSecondsUint32 := uint32(runner.setupTimeout.Seconds())
	runTimeoutSecondsUint32 := uint32(runner.testTimeout.Seconds())
	builder.WithSetupTimeoutSeconds(setupTimeoutSecondsUint32).
		WithRunTimeoutSeconds(runTimeoutSecondsUint32)
}


func (runner *AvalancheTestRunner) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	newNetwork := networksavalanche.NewAvalancheNetwork(networkCtx, runner.nodeImage)

	// first setup bootstrap nodes
	var nodeChecker []*services.DefaultAvailabilityChecker
	for i := 1; i <= runner.definedNetwork.GetNumBootstrapNodes(); i++ {
		if bootstrapNode, ok := runner.definedNetwork.Nodes[fmt.Sprintf("bootstrapNode-%d", i)]; ok {
			if bootstrapNode.IsBootstrapNode() {
				_, checker, err := newNetwork.CreateNodeNoCheck(runner.definedNetwork, bootstrapNode)
				if err != nil {
					return nil, stacktrace.Propagate(err, "An error occurred creating a new Node")
				}
				nodeChecker = append(nodeChecker, checker)
			}
		}
	}

	for _, checker := range nodeChecker {
		err := checker.WaitForStartup(15*time.Second, 10)
		if err != nil {
			panic(err)
		}
	}

	for _, node := range runner.definedNetwork.Nodes {
		if !node.IsBootstrapNode() {
			_, err := newNetwork.CreateNode(runner.definedNetwork, node)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating a new Node")
			}
		}
	}

	return newNetwork, nil
}

func (runner *AvalancheTestRunner) Run(network networks.Network) error {
	startTime := time.Now()
	if err := runner.runnableTest(network); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the test")
	}
	logrus.Infof("- - - - - - - - - - - - - - - - - - - - - Test finished in %f seconds", time.Since(startTime).Seconds())
	return nil
}
