package testsuiteavalanche

import (
	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/tests"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
)

type AvalancheTestsuite struct {
	image                 string
	datastoreServiceImage string
	isKurtosisCoreDevMode bool
}

func NewAvalancheTestsuite(avalancheImage string, isKurtosisCoreDevMode bool) *AvalancheTestsuite {
	return &AvalancheTestsuite{image: avalancheImage, isKurtosisCoreDevMode: isKurtosisCoreDevMode}
}

func (suite AvalancheTestsuite) GetTests() map[string]testsuite.Test {
	runTests := map[string]testsuite.Test{
		"PChain WorkFlow":               tests.Workflow(suite.image),
	}

	return runTests
}

func (suite AvalancheTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}
