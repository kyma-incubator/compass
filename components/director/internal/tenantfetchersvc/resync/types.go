package resync

// TenantFieldMapping missing godoc
type TenantFieldMapping struct {
	EventsField       string `envconfig:"TENANT_EVENTS_FIELD" default:"events"`

	NameField              string `envconfig:"MAPPING_FIELD_NAME" default:"name"`
	IDField                string `envconfig:"MAPPING_FIELD_ID" default:"id"`
	GlobalAccountGUIDField string `envconfig:"MAPPING_FIELD_GLOBAL_ACCOUNT_GUID" default:"globalAccountGUID"`
	SubaccountIDField      string `envconfig:"MAPPING_FIELD_SUBACCOUNT_ID" default:"subaccountId"`
	CustomerIDField        string `envconfig:"MAPPING_FIELD_CUSTOMER_ID" default:"customerId"`
	SubdomainField         string `envconfig:"MAPPING_FIELD_SUBDOMAIN" default:"subdomain"`
	DetailsField           string `envconfig:"MAPPING_FIELD_DETAILS" default:"details"`

	DiscriminatorField string `envconfig:"MAPPING_FIELD_DISCRIMINATOR"`
	DiscriminatorValue string `envconfig:"MAPPING_VALUE_DISCRIMINATOR"`

	RegionField     string `envconfig:"MAPPING_FIELD_REGION" default:"APP_MAPPING_FIELD_REGION"`
	EntityIDField   string `envconfig:"MAPPING_FIELD_ENTITY_ID" default:"entityId"`
	EntityTypeField string `envconfig:"MAPPING_FIELD_ENTITY_TYPE" default:"entityType"`

	// This is not a value from the actual event but the key under which the GlobalAccountGUIDField will be stored to avoid collisions
	GlobalAccountKey string `envconfig:"GLOBAL_ACCOUNT_KEY" default:"gaID"`
}

// MovedSubaccountsFieldMapping missing godoc
type MovedSubaccountsFieldMapping struct {
	SubaccountID string `envconfig:"MAPPING_FIELD_ID"`
	SourceTenant string `envconfig:"MOVED_SUBACCOUNT_SOURCE_TENANT_FIELD"`
	TargetTenant string `envconfig:"MOVED_SUBACCOUNT_TARGET_TENANT_FIELD"`
}

// QueryConfig contains the name of query parameters fields and default/start values
type QueryConfig struct {
	PageNumField   string `envconfig:"QUERY_PAGE_NUM_FIELD" default:"pageNum"`
	PageSizeField  string `envconfig:"QUERY_PAGE_SIZE_FIELD" default:"pageSize"`
	TimestampField string `envconfig:"QUERY_TIMESTAMP_FIELD" default:"timestamp"`
	RegionField    string `envconfig:"QUERY_REGION_FIELD" default:"region"`
	PageStartValue string `envconfig:"QUERY_PAGE_START" default:"0"`
	PageSizeValue  string `envconfig:"QUERY_PAGE_SIZE" default:"150"`
	EntityField    string `envconfig:"QUERY_ENTITY_FIELD" default:"entityId"`
}

// PageConfig missing godoc
type PageConfig struct {
	TotalPagesField   string
	TotalResultsField string
	PageNumField      string
}

type RegionalClient struct {
	RegionalAPIConfig
	RegionalClient EventAPIClient
}

// EventsType missing godoc
type EventsType int

const (
	// CreatedAccountType missing godoc
	CreatedAccountType EventsType = iota
	// DeletedAccountType missing godoc
	DeletedAccountType
	// UpdatedAccountType missing godoc
	UpdatedAccountType
	// CreatedSubaccountType missing godoc
	CreatedSubaccountType
	// DeletedSubaccountType missing godoc
	DeletedSubaccountType
	// UpdatedSubaccountType missing godoc
	UpdatedSubaccountType
	// MovedSubaccountType missing godoc
	MovedSubaccountType
)

// String missing godoc
func (e EventsType) String() string {
	switch e {
	case CreatedAccountType:
		return "CreatedEventsType"
	case DeletedAccountType:
		return "DeletedEventsType"
	case UpdatedAccountType:
		return "UpdatedEventsType"
	case CreatedSubaccountType:
		return "CreatedSubaccountType"
	case DeletedSubaccountType:
		return "DeletedSubaccountType"
	case UpdatedSubaccountType:
		return "UpdatedSubaccountType"
	case MovedSubaccountType:
		return "MovedSubaccountType"
	default:
		return ""
	}
}

// TenantEventsResponse missing godoc
type TenantEventsResponse []byte
