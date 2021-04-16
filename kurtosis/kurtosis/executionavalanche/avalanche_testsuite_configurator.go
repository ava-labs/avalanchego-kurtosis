package executionavalanche

import (
	"encoding/json"
	"strings"

	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	testsuiteAvalanche "github.com/otherview/avalanchego-kurtosis/kurtosis/kurtosis/testsuiteavalanche"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type AvalancheTestsuiteConfigurator struct{}

func NewAvalancheTestsuiteConfigurator() *AvalancheTestsuiteConfigurator {
	return &AvalancheTestsuiteConfigurator{}
}

func (t AvalancheTestsuiteConfigurator) SetLogLevel(logLevelStr string) error {
	level, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing loglevel string '%v'", logLevelStr)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return nil
}

func (t AvalancheTestsuiteConfigurator) ParseParamsAndCreateSuite(paramsJSONStr string) (testsuite.TestSuite, error) {
	paramsJSONBytes := []byte(paramsJSONStr)
	var args AvalancheTestsuiteArgs
	if err := json.Unmarshal(paramsJSONBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the testsuite params JSON")
	}

	if err := validateArgs(args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the deserialized testsuite params")
	}

	suite := testsuiteAvalanche.NewAvalancheTestsuite(args.AvalanchegoImage, args.IsKurtosisCoreDevMode)
	return suite, nil
}

func validateArgs(args AvalancheTestsuiteArgs) error {
	if strings.TrimSpace(args.AvalanchegoImage) == "" {
		return stacktrace.NewError("Avalanchego image is empty")
	}
	return nil
}
