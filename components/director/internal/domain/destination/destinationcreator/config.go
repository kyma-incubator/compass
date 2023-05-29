package destinationcreator

// Config // todo::: add godoc
type Config struct {
	BaseURL              string `envconfig:"APP_DESTINATION_BASE_URL"`
	Path                 string `envconfig:"APP_DESTINATION_PATH"`                    // todo::: "/regions/{region}/subaccounts/{subaccountId}/destinations"
	RegionParam          string `envconfig:"APP_DESTINATION_REGION_PARAMETER"`        // todo::: "region"
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_SUBACCOUNT_ID_PARAMETER"` // todo::: "subaccountId"
	DestinationNameParam string `envconfig:"APP_DESTINATION_NAME_PARAMETER"`          // todo::: "destinationName"
}
