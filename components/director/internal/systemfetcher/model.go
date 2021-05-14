package systemfetcher

type AdditionalUrls map[string]string

type AdditionalAttributes map[string]string

type System struct {
	DisplayName            string               `json:"displayName"`
	ProductDescription     string               `json:"productDescription"`
	BaseURL                string               `json:"baseUrl"`
	InfrastructureProvider string               `json:"infrastructureProvider"`
	AdditionalUrls         AdditionalUrls       `json:"additionalUrls"`
	AdditionalAttributes   AdditionalAttributes `json:"additionalAttributes"`
}
