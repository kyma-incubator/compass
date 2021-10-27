package runtime

type ServiceConfig struct {
	SelfRegisterLabelPrefix      string `envconfig:"default=clone-cmp-"`
	SelfRegisterLabel            string `envconfig:"default=cmp-xsappname"`
	SelfRegisterDistinguishLabel string `envconfig:"default=xsappname"`
	ClientID                     string `envconfig:"APP_RUNTIME_SVC_CLIENT_ID"`
	ClientSecret                 string `envconfig:"APP_RUNTIME_SVC_CLIENT_SECRET"`
	TokenURL                     string `envconfig:"APP_RUNTIME_SVC_TOKEN_URL"`
}
