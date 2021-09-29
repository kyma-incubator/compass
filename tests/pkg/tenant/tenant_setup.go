package tenant

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/tidwall/sjson"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/vrischmann/envconfig"

	"github.com/stretchr/testify/require"
)

type TenantStatus string
type TenantType string

const (
	testProvider      = "Compass Tests"
	testDefaultTenant = "Test Default"

	deleteLabelDefinitions = `DELETE FROM public.label_definitions WHERE tenant_id IN (SELECT id FROM public.business_tenant_mappings WHERE external_tenant IN (?));`

	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"

	Account  TenantType = "account"
	Customer TenantType = "customer"

	TestDefaultCustomerTenant                                  = "Test_DefaultCustomer"
	TenantSeparationTenantName                                 = "TestTenantSeparation"
	TenantsQueryNotInitializedTenantName                       = "TestTenantsQueryTenantNotInitialized"
	TenantsQueryInitializedTenantName                          = "TestTenantsQueryTenantInitialized"
	ApplicationsForRuntimeTenantName                           = "TestApplicationsForRuntime"
	ListLabelDefinitionsTenantName                             = "TestListLabelDefinitions"
	DeleteLastScenarioForApplicationTenantName                 = "TestDeleteLastScenarioForApplication"
	DeleteAutomaticScenarioAssignmentForScenarioTenantName     = "Test_DeleteAutomaticScenarioAssignmentForScenario"
	DeleteAutomaticScenarioAssignmentForSelectorTenantName     = "Test_DeleteAutomaticScenarioAssignmentForSelector"
	AutomaticScenarioAssigmentForRuntimeTenantName             = "Test_AutomaticScenarioAssigmentForRuntime"
	AutomaticScenarioAssignmentQueriesTenantName               = "Test_AutomaticScenarioAssignmentQueries"
	GetScenariosLabelDefinitionCreatesOneIfNotExistsTenantName = "TestGetScenariosLabelDefinitionCreatesOneIfNotExists"
	AutomaticScenarioAssignmentsWholeScenarioTenantName        = "TestAutomaticScenarioAssignmentsWholeScenario"
	ApplicationsForRuntimeWithHiddenAppsTenantName             = "TestApplicationsForRuntimeWithHiddenApps"
	TestDeleteApplicationIfInScenario                          = "TestDeleteApplicationIfInScenario"
)

type Tenant struct {
	ID             string       `db:"id"`
	Name           string       `db:"external_name"`
	ExternalTenant string       `db:"external_tenant"`
	ProviderName   string       `db:"provider_name"`
	Type           TenantType   `db:"type"`
	Parent         string       `db:"parent"`
	Status         TenantStatus `db:"status"`
}

type TenantIDs struct {
	TenantID               string
	SubaccountID           string
	CustomerID             string
	Subdomain              string
	SubscriptionProviderID string
}

type TenantIDProperties struct {
	TenantIDProperty               string
	SubaccountTenantIDProperty     string
	CustomerIDProperty             string
	SubdomainProperty              string
	SubscriptionProviderIDProperty string
}

var TestTenants TestTenantsManager

type TestTenantsManager struct {
	tenantsByName map[string]Tenant
}

func (mgr *TestTenantsManager) Init() {
	mgr.tenantsByName = map[string]Tenant{
		testDefaultTenant: {
			Name:           testDefaultTenant,
			ExternalTenant: "5577cf46-4f78-45fa-b55f-a42a3bdba868",
			ProviderName:   testProvider,
			Type:           Account,
			Parent:         TestDefaultCustomerTenant,
			Status:         Active,
		},
		TestDefaultCustomerTenant: {
			Name:           TestDefaultCustomerTenant,
			ExternalTenant: "2c4f4a25-ba9a-4dbc-be68-e0beb77a7eb0",
			ProviderName:   testProvider,
			Type:           Customer,
			Status:         Active,
		},
		TenantSeparationTenantName: {
			Name:           TenantSeparationTenantName,
			ExternalTenant: "f1c4b5be-b0e1-41f9-b0bc-b378200dcca0",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		ApplicationsForRuntimeTenantName: {
			Name:           ApplicationsForRuntimeTenantName,
			ExternalTenant: "5984a414-1eed-4972-af2c-b2b6a415c7d7",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		ListLabelDefinitionsTenantName: {
			Name:           ListLabelDefinitionsTenantName,
			ExternalTenant: "3f641cf5-2d14-4e0f-a122-16e7569926f1",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		DeleteLastScenarioForApplicationTenantName: {
			Name:           DeleteLastScenarioForApplicationTenantName,
			ExternalTenant: "0403be1e-f854-475e-9074-922120277af5",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		DeleteAutomaticScenarioAssignmentForScenarioTenantName: {
			Name:           DeleteAutomaticScenarioAssignmentForScenarioTenantName,
			ExternalTenant: "d08e4cb6-a77f-4a07-b021-e3317a373597",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		DeleteAutomaticScenarioAssignmentForSelectorTenantName: {
			Name:           DeleteAutomaticScenarioAssignmentForSelectorTenantName,
			ExternalTenant: "d9553135-6115-4c67-b4d9-962c00f3725f",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		AutomaticScenarioAssigmentForRuntimeTenantName: {
			Name:           AutomaticScenarioAssigmentForRuntimeTenantName,
			ExternalTenant: "8c733a45-d988-4472-af10-1256b82c70c0",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		AutomaticScenarioAssignmentQueriesTenantName: {
			Name:           AutomaticScenarioAssignmentQueriesTenantName,
			ExternalTenant: "8263cc13-5698-4a2d-9257-e8e76b543e88",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		GetScenariosLabelDefinitionCreatesOneIfNotExistsTenantName: {
			Name:           GetScenariosLabelDefinitionCreatesOneIfNotExistsTenantName,
			ExternalTenant: "2263cc13-5698-4a2d-9257-e8e76b543e33",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		AutomaticScenarioAssignmentsWholeScenarioTenantName: {
			Name:           AutomaticScenarioAssignmentsWholeScenarioTenantName,
			ExternalTenant: "65a63692-c00a-4a7d-8376-8615ee37f45c",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		ApplicationsForRuntimeWithHiddenAppsTenantName: {
			Name:           ApplicationsForRuntimeWithHiddenAppsTenantName,
			ExternalTenant: "7e1f2df8-36dc-4e40-8be3-d1555d50c91c",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		TenantsQueryInitializedTenantName: {
			Name:           TenantsQueryInitializedTenantName,
			ExternalTenant: "8cf0c909-f816-4fe3-a507-a7917ccd8380",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		TenantsQueryNotInitializedTenantName: {
			Name:           TenantsQueryNotInitializedTenantName,
			ExternalTenant: "72329135-27fd-4284-9bcb-37ea8d6307d0",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
		TestDeleteApplicationIfInScenario: {
			Name:           TestDeleteApplicationIfInScenario,
			ExternalTenant: "0d597250-6b2d-4d89-9c54-e23cb497cd01",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
	}
	mgr.Cleanup()
}

func (mgr TestTenantsManager) Cleanup() {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	transact, closeFunc, err := persistence.Configure(context.TODO(), dbCfg)

	defer func() {
		err := closeFunc()
		if err != nil {
			log.Fatal(err)
		}
	}()

	tx, err := transact.Begin()
	if err != nil {
		log.Fatal(err)
	}

	tenants := mgr.List()
	ids := make([]string, 0, len(tenants))
	for _, tnt := range tenants {
		ids = append(ids, tnt.ExternalTenant)
	}

	query, args, err := sqlx.In(deleteLabelDefinitions, ids)
	if err != nil {
		log.Fatal(err)
	}
	query = sqlx.Rebind(sqlx.BindType("postgres"), query)

	// A tenant is considered initialized if there is any labelDefinitions associated with it.
	// On first request for a given tenant a labelDefinition for key scenario and value DEFAULT is created.
	// Therefore once accessed a tenant is considered initialized. That's the reason we clean up (uninitialize) all the tests tenants here.
	// There is a test relying on this (testing tenants graphql query).
	_, err = tx.ExecContext(context.TODO(), query, args...)
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (mgr TestTenantsManager) GetIDByName(t require.TestingT, name string) string {
	val, ok := mgr.tenantsByName[name]
	require.True(t, ok)
	return val.ExternalTenant
}

func (mgr TestTenantsManager) GetDefaultTenantID() string {
	return mgr.tenantsByName[testDefaultTenant].ExternalTenant
}

func (mgr TestTenantsManager) EmptyTenant() string {
	return ""
}

func (mgr TestTenantsManager) List() []Tenant {
	var toReturn []Tenant

	for _, v := range mgr.tenantsByName {
		toReturn = append(toReturn, v)
	}

	return toReturn
}

// CreateTenantRequest returns a prepared tenant request with token in the header with the necessary tenant-fetcher claims
func CreateTenantRequest(t *testing.T, tenantIDs TenantIDs, tenantProperties TenantIDProperties, httpMethod, tenantFetcherUrl, externalServicesMockURL string) *http.Request {
	var (
		body = "{}"
		err  error
	)

	if len(tenantIDs.TenantID) > 0 {
		body, err = sjson.Set(body, tenantProperties.TenantIDProperty, tenantIDs.TenantID)
		require.NoError(t, err)
	}
	if len(tenantIDs.SubaccountID) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubaccountTenantIDProperty, tenantIDs.SubaccountID)
		require.NoError(t, err)
	}
	if len(tenantIDs.CustomerID) > 0 {
		body, err = sjson.Set(body, tenantProperties.CustomerIDProperty, tenantIDs.CustomerID)
		require.NoError(t, err)
	}
	if len(tenantIDs.Subdomain) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubdomainProperty, tenantIDs.Subdomain)
		require.NoError(t, err)
	}
	if len(tenantIDs.SubscriptionProviderID) > 0 {
		body, err = sjson.Set(body, tenantProperties.SubscriptionProviderIDProperty, tenantIDs.SubscriptionProviderID)
		require.NoError(t, err)
	}

	request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
	require.NoError(t, err)
	claims := map[string]interface{}{
		"test": "tenant-fetcher",
		"scope": []string{
			"prefix.Callback",
		},
		"tenant":   "tenant",
		"identity": "tenant-fetcher-tests",
		"iss":      externalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, externalServicesMockURL, claims)))

	return request
}

func ActualTenantID(tenantIDs TenantIDs) string {
	if len(tenantIDs.SubaccountID) > 0 {
		return tenantIDs.SubaccountID
	}

	return tenantIDs.TenantID
}