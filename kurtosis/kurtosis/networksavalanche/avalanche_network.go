// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package networksavalanche

import (
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/kurtosis/servicesavalanche/avalanchegonode"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
)

const (
	waitForStartupTimeBetweenPolls = 20 * time.Second
	waitForStartupMaxNumPolls      = 10
	waitForTermination             = 30 * time.Second
)

type AvalancheNetwork struct {
	networkCtx      *networks.NetworkContext
	apiServiceImage string
	nodes           map[services.ServiceID]*avalanchegonode.NodeAPIService
}

func NewAvalancheNetwork(networkCtx *networks.NetworkContext, apiServiceImage string) *AvalancheNetwork {
	return &AvalancheNetwork{
		networkCtx:      networkCtx,
		apiServiceImage: apiServiceImage,
		nodes:           map[services.ServiceID]*avalanchegonode.NodeAPIService{},
	}
}

func Cast(network networks.Network, context *testsuite.TestContext) *AvalancheNetwork {
	avalancheNetwork, ok := network.(*AvalancheNetwork)
	if !ok {
		context.Fatal(stacktrace.NewError("network not AvalancheNetwork type"))
	}
	return avalancheNetwork
}

func (network *AvalancheNetwork) CreateNodeNoCheck(definedNetwork *networkbuilder.Network, node *networkbuilder.Node) (services.ServiceID, *services.DefaultAvailabilityChecker, error) {
	serviceID := services.ServiceID(node.ID)
	if _, ok := network.nodes[serviceID]; ok {
		return serviceID, nil, fmt.Errorf("node with the same nodeID already exists")
	}

	initializer := avalanchegonode.NewNodeInitializer(definedNetwork, node, network.nodes)
	uncastedService, checker, err := network.networkCtx.AddService(serviceID, initializer)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	castedService := uncastedService.(*avalanchegonode.NodeAPIService)
	network.nodes[serviceID] = castedService
	return serviceID, checker.(*services.DefaultAvailabilityChecker), nil
}

func (network *AvalancheNetwork) CreateNode(definedNetwork *networkbuilder.Network, node *networkbuilder.Node) (services.ServiceID, error) {
	serviceID := services.ServiceID(node.ID)
	if _, ok := network.nodes[serviceID]; ok {
		return serviceID, fmt.Errorf("node with the same nodeID already exists")
	}

	initializer := avalanchegonode.NewNodeInitializer(definedNetwork, node, network.nodes)
	uncastedService, checker, err := network.networkCtx.AddService(serviceID, initializer)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred adding the API service")
	}
	if err = checker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred waiting for the API service to start")
	}
	castedService := uncastedService.(*avalanchegonode.NodeAPIService)
	network.nodes[serviceID] = castedService
	return serviceID, nil
}

func (network *AvalancheNetwork) GetNodeClient(nodeID string) (*avalanchegoclient.Client, error) {
	serviceID := services.ServiceID(nodeID)
	service, found := network.nodes[serviceID]
	if !found {
		return nil, stacktrace.NewError("No API service with ID '%v' has been added", serviceID)
	}

	return service.GetNodeClient(), nil
}

func (network *AvalancheNetwork) GetClient() string {
	for _, service := range network.nodes {
		if service != nil {
			ip := service.GetIPAddress()
			fmt.Printf("returning IP address - %v\n", ip)
			return ip
		}
	}
	return ""
}

func (network *AvalancheNetwork) GetIPAddress(nodeID string) (string, error) {
	serviceID := services.ServiceID(nodeID)
	service, found := network.nodes[serviceID]
	if !found {
		return "", stacktrace.NewError("No node service with ID '%v' has been added", serviceID)
	}

	return service.GetIPAddress(), nil
}

func (network *AvalancheNetwork) RemoveNode(definedNetwork *networkbuilder.Network, node *networkbuilder.Node) error {
	serviceID := services.ServiceID(node.ID)
	if _, ok := network.nodes[serviceID]; !ok {
		return fmt.Errorf("node does not exist in the defined services")
	}

	definedNetwork.RemoveNode(node)

	return network.networkCtx.RemoveService(serviceID, uint64(waitForTermination.Seconds()))
}
