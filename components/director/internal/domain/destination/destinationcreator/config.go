package destinationcreator

// Config // todo::: add godoc
type Config struct {
	BaseURL              string `envconfig:"APP_DESTINATION_BASE_URL"`
	Path                 string `envconfig:"APP_DESTINATION_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_REGION_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_SUBACCOUNT_ID_PARAMETER"`
	DestinationNameParam string `envconfig:"APP_DESTINATION_NAME_PARAMETER"`
}
