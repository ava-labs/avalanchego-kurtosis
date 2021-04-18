// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanchegonode

import (
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/avalanchegoclient"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/sirupsen/logrus"
)

type NodeAPIService struct {
	serviceID          services.ServiceID
	ipAddr             string
	httpPort           int
	stakingPort        int
	bootstrappedPChain bool
	bootstrappedCChain bool
	bootstrappedXChain bool
}

func NewNodeAPIService(serviceID services.ServiceID, ipAddr string, httpPort int, stakePort int) *NodeAPIService {
	return &NodeAPIService{serviceID: serviceID, ipAddr: ipAddr, httpPort: httpPort, stakingPort: stakePort}
}

// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
func (service *NodeAPIService) GetServiceID() services.ServiceID {
	return service.serviceID
}

func (service *NodeAPIService) GetIPAddress() string {
	return service.ipAddr
}

func (service *NodeAPIService) IsAvailable() bool {
	checkClient := avalanchegoclient.NewClient(service.ipAddr, service.httpPort, 10*time.Second)

	logrus.Infof("Node: %s -> Bootstrapped P: %v Bootstrapped C: %v Bootstrapped X: %v\n",
		service.serviceID,
		service.bootstrappedPChain,
		service.bootstrappedCChain,
		service.bootstrappedXChain,
	)

	if !service.bootstrappedPChain {
		if bootstrapped, err := checkClient.InfoAPI().IsBootstrapped("P"); err != nil || !bootstrapped {
			return false
		}
		service.bootstrappedPChain = true
	}

	if !service.bootstrappedCChain {
		if bootstrapped, err := checkClient.InfoAPI().IsBootstrapped("C"); err != nil || !bootstrapped {
			return false
		}
		service.bootstrappedCChain = true
	}

	if !service.bootstrappedXChain {
		if bootstrapped, err := checkClient.InfoAPI().IsBootstrapped("X"); err != nil || !bootstrapped {
			return false
		}
		service.bootstrappedXChain = true
	}

	// todo we should use the health api
	bootstrapped := service.bootstrappedPChain && service.bootstrappedCChain && service.bootstrappedXChain
	if bootstrapped {
		logrus.Infof("Node: %s is bootstrapped", service.serviceID)
	}

	return bootstrapped
}

// ===========================================================================================
//                         API service-specific methods
// ===========================================================================================

func (service *NodeAPIService) GetNodeClient() *avalanchegoclient.Client {
	return avalanchegoclient.NewClient(service.ipAddr, service.httpPort, 10*time.Second)
}

func (service *NodeAPIService) GetStakingPort() int {
	return service.stakingPort
}

func (service *NodeAPIService) GetHTTPPort() int {
	return service.httpPort
}
