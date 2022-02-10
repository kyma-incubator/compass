package config

type AuditlogConfig struct {
	TokenURL          string `envconfig:"AUDITLOG_TOKEN_URL"`
	ClientID          string `envconfig:"AUDITLOG_CLIENT_ID"`
	X509Cert          string `envconfig:"AUDITLOG_X509_CERT"`
	X509Key           string `envconfig:"AUDITLOG_X509_KEY"`
	ManagementURL     string `envconfig:"AUDITLOG_MANAGEMENT_URL"`
	ManagementAPIPath string `envconfig:"AUDITLOG_MANAGEMENT_API_PATH"`
}
