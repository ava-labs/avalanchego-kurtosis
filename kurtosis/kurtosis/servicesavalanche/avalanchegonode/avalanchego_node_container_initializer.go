package avalanchegonode

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
)

const (
	testVolumeMountpoint = "/test-volume"
	configFileID         = "cChainConfig"
	configFileContents   = `{"coreth-config":{"snowman-api-enabled": false,"coreth-admin-api-enabled": false,"net-api-enabled": true,"rpc-gas-cap": 2500000000,"rpc-tx-fee-cap": 100,"eth-api-enabled": true,"personal-api-enabled": true,"tx-pool-api-enabled": true,"debug-api-enabled": false,"web3-api-enabled": true,"local-txs-enabled": true}}`
)

type NodeInitializer struct {
	node             *networkbuilder.Node
	definedNetwork   *networkbuilder.Network
	createdNodes     map[services.ServiceID]*NodeAPIService
	bootstrapNodes   string
	bootstrapNodeIPs string
}

func NewNodeInitializer(definedNetwork *networkbuilder.Network, node *networkbuilder.Node, nodes map[services.ServiceID]*NodeAPIService) *NodeInitializer {
	return &NodeInitializer{definedNetwork: definedNetwork, node: node, createdNodes: nodes}
}

func (initializer *NodeInitializer) GetDockerImage() string {
	return initializer.node.GetImage()
}

func (initializer *NodeInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		fmt.Sprintf("%v/tcp", initializer.node.GetStakingPort()): true,
		fmt.Sprintf("%v/tcp", initializer.node.GetHTTPPort()):    true,
	}
}

func (initializer *NodeInitializer) GetServiceWrappingFunc() func(serviceId services.ServiceID, ipAddr string) services.Service {
	return func(serviceId services.ServiceID, ipAddr string) services.Service {
		return NewNodeAPIService(serviceId, ipAddr, initializer.node.GetHTTPPort(), initializer.node.GetStakingPort())
	}
}

func (initializer *NodeInitializer) GetFilesToGenerate() map[string]bool {
	files := map[string]bool{
		configFileID: true,
	}

	if initializer.node.HasCerts() {
		files[constants.StakingTLSCertFileID] = true
		files[constants.StakingTLSKeyFileID] = true
	}
	return files
}

func (initializer *NodeInitializer) InitializeGeneratedFiles(mountedFiles map[string]*os.File) error {
	configFilePointer := mountedFiles[configFileID]
	if _, err := configFilePointer.WriteString(configFileContents); err != nil {
		return err
	}

	if initializer.node.HasCerts() {
		certFilePointer := mountedFiles[constants.StakingTLSCertFileID]
		keyFilePointer := mountedFiles[constants.StakingTLSKeyFileID]
		if _, err := certFilePointer.Write(bytes.NewBufferString(initializer.node.GetTLSCert()).Bytes()); err != nil {
			return err
		}
		if _, err := keyFilePointer.Write([]byte(initializer.node.GetPrivateKey())); err != nil {
			return err
		}
	}
	return nil
}

func (initializer *NodeInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	return map[services.FilesArtifactID]string{}
}

func (initializer *NodeInitializer) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func (initializer *NodeInitializer) GetStartCommand(mountedFileFilepaths map[string]string, ipPlaceholder string) ([]string, error) {
	publicIPFlag := fmt.Sprintf("--public-ip=%s", ipPlaceholder)
	commandList := []string{
		"/avalanchego/build/avalanchego",
		publicIPFlag,
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", 9650),
		"--http-host=", // Leave empty to make API openly accessible
		fmt.Sprintf("--staking-port=%d", 9651),
		fmt.Sprintf("--log-level=%s", "debug"),
		fmt.Sprintf("--snow-sample-size=%d", initializer.definedNetwork.GetSnowSampleSize()),
		fmt.Sprintf("--snow-quorum-size=%d", initializer.definedNetwork.GetSnowQuorumSize()),
		fmt.Sprintf("--staking-enabled=%v", initializer.node.GetStaking()),
		fmt.Sprintf("--tx-fee=%d", initializer.definedNetwork.GetTxFee()),
	}

	if initializer.node.HasCerts() {
		commandList = append(commandList, fmt.Sprintf("--staking-tls-cert-file=\"%s\"", mountedFileFilepaths[constants.StakingTLSCertFileID]))
		commandList = append(commandList, fmt.Sprintf("--staking-tls-key-file=\"%s\"", mountedFileFilepaths[constants.StakingTLSKeyFileID]))
	}

	// NOTE: This seems weird, BUT there's a reason for it: An avalanche node doesn't use certs, and instead relies on
	//  the user explicitly passing in the node ID of the bootstrapper it wants. This prevents man-in-the-middle
	//  attacks, just like using a cert would. Us hardcoding this bootstrapper ID here is the equivalent
	//  of a user knowing the node ID in advance, which provides the same level of protection.
	commandList = append(commandList, "--bootstrap-ids="+initializer.node.GetConnectedBTNodeIDs())

	// find the ips of the bootstrap nodes
	bootstrapNodeID := initializer.node.GetBootstrapNodeID()
	var joinedBootStrapNodes []string
	var bootstrapIPs string
	if bootstrapNodeID == 0 {
		bootstrapNodeID = initializer.definedNetwork.GetNumBootstrapNodes() + 1
	}

	for i := 1; i < bootstrapNodeID; i++ {
		otherBootstrapNode, ok := initializer.createdNodes[services.ServiceID(fmt.Sprintf("bootstrapNode-%d", i))]
		if !ok {
			panic(fmt.Sprintf("trying to address a bootstrap-%d node that does not exist", i))
		}
		joinedBootStrapNodes = append(joinedBootStrapNodes, fmt.Sprintf("%s:%d", otherBootstrapNode.GetIPAddress(), otherBootstrapNode.GetStakingPort()))
	}
	bootstrapIPs = strings.Join(joinedBootStrapNodes, ",")

	commandList = append(commandList, "--bootstrap-ips="+bootstrapIPs)

	// Create new command list that adds suffix to config file (otherwise viper
	// cannot open it)
	configFilepath, found := mountedFileFilepaths[configFileID]
	if found {
		commandList = append(commandList, fmt.Sprintf("--config-file=\"%s.json\"", configFilepath))
		combinedCommandList := strings.Join(commandList, " ")
		commandList = []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("mv \"%s\" \"%s.json\" && %s", configFilepath, configFilepath, combinedCommandList),
		}
	}

	logrus.Infof("Command list for node: %v -> %v\n", initializer.node, commandList)
	return commandList, nil
}
