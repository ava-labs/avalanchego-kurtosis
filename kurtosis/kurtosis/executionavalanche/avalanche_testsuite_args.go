// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executionavalanche

type AvalancheTestsuiteArgs struct {
	AvalanchegoImage          string `json:"avalanchegoImage"`

	// Indicates that this testsuite is being run as part of CI testing in Kurtosis Core
	IsKurtosisCoreDevMode bool `json:"isKurtosisCoreDevMode"`
}
