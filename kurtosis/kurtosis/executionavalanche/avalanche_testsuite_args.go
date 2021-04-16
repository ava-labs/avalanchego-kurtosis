package executionavalanche

type AvalancheTestsuiteArgs struct {
	AvalanchegoImage          string `json:"avalanchegoImage"`

	// Indicates that this testsuite is being run as part of CI testing in Kurtosis Core
	IsKurtosisCoreDevMode bool `json:"isKurtosisCoreDevMode"`
}
