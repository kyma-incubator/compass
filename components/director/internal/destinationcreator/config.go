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
	SubaccountLevelPath  string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_PATH"`
	InstanceLevelPath    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_INSTANCE_LEVEL_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_REGION_PARAMETER"`
	InstanceIDParam      string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_INSTANCE_ID_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_SUBACCOUNT_ID_PARAMETER"`
	DestinationNameParam string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_NAME_PARAMETER"`
}

// CertificateAPIConfig holds a configuration specific for the certificate API of the destination creator service
type CertificateAPIConfig struct {
	BaseURL              string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_BASE_URL"`
	SubaccountLevelPath  string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_PATH"`
	InstanceLevelPath    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_INSTANCE_LEVEL_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_REGION_PARAMETER"`
	InstanceIDParam      string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_INSTANCE_ID_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_SUBACCOUNT_ID_PARAMETER"`
	CertificateNameParam string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_NAME_PARAMETER"`
	FileNameKey          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_FILE_NAME_KEY"`
	CommonNameKey        string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_COMMON_NAME_KEY"`
	CertificateChainKey  string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_CERT_CHAIN_KEY"`
}

// URLConfig contains URL configuration properties
type URLConfig struct {
	BaseURL             string
	SubaccountLevelPath string
	InstanceLevelPath   string
	RegionParam         string
	NameParam           string
	SubaccountIDParam   string
	InstanceIDParam     string
}

// URLParameters contains URL path parameters configuration
type URLParameters struct {
	EntityName   string
	Region       string
	SubaccountID string
	InstanceID   string
}
