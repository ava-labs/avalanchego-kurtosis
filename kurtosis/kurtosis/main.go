//
// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/execution"
	"github.com/sirupsen/logrus"

	executionAvalanche "github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/executionavalanche"
	ethLog "github.com/ethereum/go-ethereum/log"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {
	customParamsJSONArg := flag.String(
		"custom-params-json",
		"{}",
		"JSON string containing custom data that the testsuite will deserialize to modify runtime behaviour",
	)

	kurtosisAPISocketArg := flag.String(
		"kurtosis-api-socket",
		"",
		"Socket in the form of address:port of the Kurtosis API container",
	)

	logLevelArg := flag.String(
		"log-level",
		"",
		"String indicating the loglevel that the test suite should output with",
	)

	flag.Parse()

	// avalanche test suit configurator
	configurator := executionAvalanche.NewAvalancheTestsuiteConfigurator()
	if err := configurator.SetLogLevel(logrus.DebugLevel.String()); err != nil {
		os.Exit(failureExitCode)
	}
	ethLogLevel, err := ethLog.LvlFromString(*logLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the ethLog log level string: %v\n", err)
		os.Exit(1)
	}
	ethLog.Root().SetHandler(ethLog.LvlFilterHandler(ethLogLevel, ethLog.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	suiteExecutor := execution.NewTestSuiteExecutor(*kurtosisAPISocketArg, *logLevelArg, *customParamsJSONArg, configurator)
	if err := suiteExecutor.Run(context.Background()); err != nil {
		logrus.Errorf("An error occurred running the test suite executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
