package avalanchegonode

import (
	"bytes"
	"fmt"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	testVolumeMountpoint = "/test-volume"
	configFileID         = "cChainConfig"
	configFileContents   = `{"coreth-config":{"snowman-api-enabled": false,"coreth-admin-api-enabled": false,"net-api-enabled": true,"rpc-gas-cap": 2500000000,"rpc-tx-fee-cap": 100,"eth-api-enabled": true,"personal-api-enabled": true,"tx-pool-api-enabled": true,"debug-api-enabled": false,"web3-api-enabled": true,"local-txs-enabled": true}}`
)

type AvalancheGoContainerConfigFactory struct {
	nodeConfig             *networkbuilder.Node
	definedNetwork   *networkbuilder.Network
	createdNodes     map[services.ServiceID]*NodeAPIService
	bootstrapNodes   string
	bootstrapNodeIPs string
}

func NewAvalancheGoContainerConfigFactory(definedNetwork *networkbuilder.Network, nodeConfig *networkbuilder.Node, nodes map[services.ServiceID]*NodeAPIService) *AvalancheGoContainerConfigFactory {
	return &AvalancheGoContainerConfigFactory{definedNetwork: definedNetwork, nodeConfig: nodeConfig, createdNodes: nodes}
}

func (factory AvalancheGoContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	serviceCreatingFunc := func(serviceCtx *services.ServiceContext) services.Service {
		return NewNodeAPIService(serviceCtx, factory.nodeConfig.GetHTTPPort(), factory.nodeConfig.GetStakingPort())
	}

	fileGeneratingFuncs := map[string]func(*os.File) error{
		configFileID: func(fp *os.File) error {
			if _, err := fp.WriteString(configFileContents); err != nil {
				return stacktrace.Propagate(err, "An error occurred writing config file contents to fp")
			}
			return nil
		},
	}
	if factory.nodeConfig.HasCerts() {
		fileGeneratingFuncs[constants.StakingTLSCertFileID] = func(fp *os.File) error {
			if _, err := fp.Write(bytes.NewBufferString(factory.nodeConfig.GetTLSCert()).Bytes()); err != nil {
				return stacktrace.Propagate(err, "An error occurred writing the TLS cert file to fp")
			}
			return nil
		}

		fileGeneratingFuncs[constants.StakingTLSKeyFileID] = func(fp *os.File) error {
			if _, err := fp.Write([]byte(factory.nodeConfig.GetPrivateKey())); err != nil {
				return stacktrace.Propagate(err, "An error occurred writing the private key to fp")
			}
			return nil
		}
	}

	result := services.NewContainerCreationConfigBuilder(factory.nodeConfig.GetImage(), testVolumeMountpoint, serviceCreatingFunc).
		WithUsedPorts(map[string]bool{
			fmt.Sprintf("%v/tcp", factory.nodeConfig.GetStakingPort()): true,
			fmt.Sprintf("%v/tcp", factory.nodeConfig.GetHTTPPort()):    true,
		}).
		WithGeneratedFiles(fileGeneratingFuncs).
		Build()
	return result, nil
}

func (factory AvalancheGoContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*services.ContainerRunConfig, error) {
	publicIPFlag := fmt.Sprintf("--public-ip=%s", containerIpAddr)
	commandList := []string{
		"/avalanchego/build/avalanchego",
		publicIPFlag,
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", 9650),
		"--http-host=", // Leave empty to make API openly accessible
		fmt.Sprintf("--staking-port=%d", 9651),
		fmt.Sprintf("--log-level=%s", "debug"),
		fmt.Sprintf("--snow-sample-size=%d", factory.definedNetwork.GetSnowSampleSize()),
		fmt.Sprintf("--snow-quorum-size=%d", factory.definedNetwork.GetSnowQuorumSize()),
		fmt.Sprintf("--staking-enabled=%v", factory.nodeConfig.GetStaking()),
		fmt.Sprintf("--tx-fee=%d", factory.definedNetwork.GetTxFee()),
	}

	if factory.nodeConfig.HasCerts() {
		commandList = append(commandList, fmt.Sprintf("--staking-tls-cert-file=\"%s\"", generatedFileFilepaths[constants.StakingTLSCertFileID]))
		commandList = append(commandList, fmt.Sprintf("--staking-tls-key-file=\"%s\"", generatedFileFilepaths[constants.StakingTLSKeyFileID]))
	}

	// NOTE: This seems weird, BUT there's a reason for it: An avalanche node doesn't use certs, and instead relies on
	//  the user explicitly passing in the node ID of the bootstrapper it wants. This prevents man-in-the-middle
	//  attacks, just like using a cert would. Us hardcoding this bootstrapper ID here is the equivalent
	//  of a user knowing the node ID in advance, which provides the same level of protection.
	commandList = append(commandList, "--bootstrap-ids="+factory.nodeConfig.GetConnectedBTNodeIDs())

	// find the ips of the bootstrap nodes
	bootstrapNodeID := factory.nodeConfig.GetBootstrapNodeID()
	var joinedBootStrapNodes []string
	var bootstrapIPs string
	if bootstrapNodeID == 0 {
		bootstrapNodeID = factory.definedNetwork.GetNumBootstrapNodes() + 1
	}

	for i := 1; i < bootstrapNodeID; i++ {
		otherBootstrapNode, ok := factory.createdNodes[services.ServiceID(fmt.Sprintf("bootstrapNode-%d", i))]
		if !ok {
			panic(fmt.Sprintf("trying to address a bootstrap-%d node that does not exist", i))
		}
		joinedBootStrapNodes = append(joinedBootStrapNodes, fmt.Sprintf("%s:%d", otherBootstrapNode.GetIPAddress(), otherBootstrapNode.GetStakingPort()))
	}
	bootstrapIPs = strings.Join(joinedBootStrapNodes, ",")

	commandList = append(commandList, "--bootstrap-ips="+bootstrapIPs)

	// Create new command list that adds suffix to config file (otherwise viper
	// cannot open it)
	configFilepath, found := generatedFileFilepaths[configFileID]
	if found {
		commandList = append(commandList, fmt.Sprintf("--config-file=\"%s.json\"", configFilepath))
		combinedCommandList := strings.Join(commandList, " ")
		commandList = []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("mv \"%s\" \"%s.json\" && %s", configFilepath, configFilepath, combinedCommandList),
		}
	}

	logrus.Infof("Command list for node: %v -> %v\n", factory.nodeConfig, commandList)

	result := services.NewContainerRunConfigBuilder().
		WithCmdOverride(commandList).
		Build()

	return result, nil
}