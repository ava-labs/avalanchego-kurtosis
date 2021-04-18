// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package runner

import (
	"fmt"
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type AvalancheTestRunner struct {
	nodeImage      string
	definedNetwork *networkbuilder.Network
	runnableTest   func(network networks.Network, context testsuite.TestContext)
	testTimeout    time.Duration
	setupTimeout   time.Duration
}

func NewGenericAvalancheTestRunner(definedNetwork *networkbuilder.Network, test func(network networks.Network, context testsuite.TestContext), testTimeout time.Duration, setupTimeout time.Duration) *AvalancheTestRunner {
	return &AvalancheTestRunner{
		definedNetwork: definedNetwork,
		runnableTest:   test,
		testTimeout:    testTimeout,
		setupTimeout:   setupTimeout,
	}
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

func (runner *AvalancheTestRunner) Run(network networks.Network, testCtx testsuite.TestContext) {
	startTime := time.Now()
	runner.runnableTest(
		network,
		testCtx,
	)
	logrus.Infof("- - - - - - - - - - - - - - - - - - - - - Test finished in %f seconds", time.Since(startTime).Seconds())
}

func (runner *AvalancheTestRunner) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{}
}

func (runner *AvalancheTestRunner) GetExecutionTimeout() time.Duration {
	return runner.testTimeout
}

func (runner *AvalancheTestRunner) GetSetupTimeout() time.Duration {
	return runner.setupTimeout
}
