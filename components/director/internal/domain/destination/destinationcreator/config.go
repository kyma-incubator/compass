package destinationcreator

// Config holds destination creator service API configuration
type Config struct {
	CorrelationIDsKey string `envconfig:"APP_DESTINATION_CREATOR_CORRELATION_IDS_KEY"`
	*DestinationAPIConfig
	*CertificateAPIConfig
}

// DestinationAPIConfig holds a configuration specific for the destination API of the destination creator service
type DestinationAPIConfig struct {
	BaseURL              string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_BASE_URL"`
	Path                 string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_REGION_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_SUBACCOUNT_ID_PARAMETER"`
	DestinationNameParam string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_NAME_PARAMETER"`
}

// CertificateAPIConfig holds a configuration specific for the certificate API of the destination creator service
type CertificateAPIConfig struct {
	BaseURL              string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_BASE_URL"`
	Path                 string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_REGION_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_SUBACCOUNT_ID_PARAMETER"`
	CertificateNameParam string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_NAME_PARAMETER"`
	FileNameKey          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_FILE_NAME_KEY"`
	CommonNameKey        string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_COMMON_NAME_KEY"`
	CertificateChainKey  string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_CERT_CHAIN_KEY"`
}
