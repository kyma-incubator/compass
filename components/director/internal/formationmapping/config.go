package formationmapping

// Config holds the configuration available for the formation mapping
type Config struct {
	AsyncAPIPathPrefix                             string `envconfig:"APP_FORMATION_MAPPING_API_PATH_PREFIX"`
	AsyncFormationAssignmentStatusAPIEndpoint      string `envconfig:"APP_FORMATION_ASSIGNMENT_ASYNC_STATUS_API_ENDPOINT"`
	AsyncFormationAssignmentStatusResetAPIEndpoint string `envconfig:"APP_FORMATION_ASSIGNMENT_ASYNC_STATUS_RESET_API_ENDPOINT"`
	AsyncFormationStatusAPIEndpoint                string `envconfig:"APP_FORMATION_ASYNC_STATUS_API_ENDPOINT"`
	UCLCertOUSubaccountID                          string `envconfig:"APP_UCL_CERT_OU_SUBACCOUNT_ID"`
}
