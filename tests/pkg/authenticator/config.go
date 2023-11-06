package authenticator

type Config struct {
	Account    AccountConfig
	Subaccount SubaccountConfig
}

type AccountConfig struct {
	TokenURL       string `envconfig:"APP_USER_AUTHENTICATOR_ACCOUNT_TOKEN_URL"`
	ClientID       string `envconfig:"APP_USER_AUTHENTICATOR_ACCOUNT_CLIENT_ID"`
	ClientSecret   string `envconfig:"APP_USER_AUTHENTICATOR_ACCOUNT_CLIENT_SECRET"`
	OAuthTokenPath string `envconfig:"APP_USER_AUTHENTICATOR_ACCOUNT_TOKEN_PATH"`
	Subdomain      string `envconfig:"APP_USER_AUTHENTICATOR_ACCOUNT_SUBDOMAIN"`
}

type SubaccountConfig struct {
	TokenURL       string `envconfig:"APP_USER_AUTHENTICATOR_SUBACCOUNT_TOKEN_URL"`
	ClientID       string `envconfig:"APP_USER_AUTHENTICATOR_SUBACCOUNT_CLIENT_ID"`
	ClientSecret   string `envconfig:"APP_USER_AUTHENTICATOR_SUBACCOUNT_CLIENT_SECRET"`
	OAuthTokenPath string `envconfig:"APP_USER_AUTHENTICATOR_SUBACCOUNT_TOKEN_PATH"`
	Subdomain      string `envconfig:"APP_USER_AUTHENTICATOR_SUBACCOUNT_SUBDOMAIN"`
}
