package systemfetcher

type Capabilities []string

type AdditionalUrls map[string]string

type AdditionalAttributes map[string]string

type ProductInstanceExtended struct {
	ID                             string               `json:"id"`
	DisplayName                    string               `json:"displayName"`
	ProductID                      string               `json:"productId"`
	ProductDescription             string               `json:"productDescription"`
	ZoneID                         string               `json:"zoneId"`
	SystemID                       string               `json:"systemId"`
	HomepageURL                    string               `json:"homepageUrl"`
	BaseURL                        string               `json:"baseUrl"`
	BaseAPIURL                     string               `json:"baseAPIUrl"`
	Capabilities                   Capabilities         `json:"capabilities"`
	InfrastructureProvider         string               `json:"infrastructureProvider"`
	RegionID                       string               `json:"regionId"`
	LastChangeDateTime             string               `json:"lastChangeDateTime"`
	ProductInstanceType            string               `json:"productInstanceType"`
	ExternalID                     string               `json:"externalId"`
	SystemExternalID               string               `json:"systemExternalId"`
	SystemNumber                   string               `json:"systemNumber"`
	LogicalSystemID                string               `json:"logicalSystemId"`
	LogicalSystemName              string               `json:"logicalSystemName"`
	PPMSProductID                  string               `json:"ppmsProductId"`
	AdditionalUrls                 AdditionalUrls       `json:"additionalUrls"`
	PPMSProductLineID              string               `json:"ppmsProductLineId"`
	PPMSProductLineOfficialName    string               `json:"ppmsProductLineOfficialName"`
	PPMSProductVersionID           string               `json:"ppmsProductVersionId"`
	PPMSProductVersionName         string               `json:"ppmsProductVersionName"`
	PPMSProductVersionOfficialName string               `json:"ppmsProductVersionOfficialName"`
	PPMSProductVersionRelease      string               `json:"ppmsProductVersionRelease"`
	CRMCustomerID                  string               `json:"crmCustomerId"`
	ERPCustomerID                  string               `json:"erpCustomerId"`
	CustomerName                   string               `json:"customerName"`
	RegionName                     string               `json:"regionName"`
	Role                           string               `json:"role"`
	InstallationNumber             string               `json:"installationNumber"`
	SystemName                     string               `json:"systemName"`
	SystemIDSAP                    string               `json:"systemIdSap"`
	SystemNameSAPDefault           string               `json:"systemNameSapDefault"`
	AdditionalAttributes           AdditionalAttributes `json:"additionalAttributes"`
}
