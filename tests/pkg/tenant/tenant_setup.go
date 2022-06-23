package tenant

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type TenantStatus string
type TenantType string

const (
	testProvider      = "Compass Tests"
	testDefaultTenant = "Test Default"

	deleteLabelDefinitionsQuery = `DELETE FROM public.label_definitions WHERE tenant_id IN (SELECT id FROM public.business_tenant_mappings WHERE external_tenant IN (?));`
	deleteFormationsQuery       = `DELETE FROM public.formations WHERE tenant_id IN (SELECT id FROM public.business_tenant_mappings WHERE external_tenant IN (?));`

	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"

	Account    TenantType = "account"
	Customer   TenantType = "customer"
	Subaccount TenantType = "subaccount"

	TestDefaultCustomerTenant                                  = "Test_DefaultCustomer"
	TestSystemFetcherTenant                                    = "TestSystemFetcherAccount"
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
	TestProviderSubaccount                                     = "TestProviderSubaccount"
	TestConsumerSubaccount                                     = "TestConsumerSubaccount"
	TestIntegrationSystemManagedSubaccount                     = "TestIntegrationSystemManagedSubaccount"
	TestIntegrationSystemManagedAccount                        = "TestIntegrationSystemManagedAccount"
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
		TestSystemFetcherTenant: {
			Name:           TestSystemFetcherTenant,
			ExternalTenant: "c395681d-11dd-4cde-bbcf-570b4a153e79",
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
		TestProviderSubaccount: {
			Name:           TestProviderSubaccount,
			ExternalTenant: "f8075207-1478-4a80-bd26-24a4785a2bfd",
			ProviderName:   testProvider,
			Type:           Subaccount,
			Status:         Active,
			Parent:         testDefaultTenant,
		},
		TestConsumerSubaccount: {
			Name:           TestConsumerSubaccount,
			ExternalTenant: "1f538f34-30bf-4d3d-aeaa-02e69eef84ae",
			ProviderName:   testProvider,
			Type:           Subaccount,
			Status:         Active,
			Parent:         ApplicationsForRuntimeTenantName,
		},
		TestIntegrationSystemManagedSubaccount: {
			Name:           TestIntegrationSystemManagedSubaccount,
			ExternalTenant: "3cfcdd62-320d-403b-b66a-4ee3cdd06947",
			ProviderName:   testProvider,
			Type:           Subaccount,
			Status:         Active,
			Parent:         testDefaultTenant,
		},
		TestIntegrationSystemManagedAccount: {
			Name:           TestIntegrationSystemManagedAccount,
			ExternalTenant: "7e8ab2e3-3bb4-42e3-92b2-4e0bf48559d3",
			ProviderName:   testProvider,
			Type:           Account,
			Status:         Active,
		},
	}
	mgr.Cleanup()
}

func (mgr TestTenantsManager) Cleanup() {
	tenants := mgr.List()
	ids := make([]string, 0, len(tenants))
	for _, tnt := range tenants {
		ids = append(ids, tnt.ExternalTenant)
	}
	mgr.cleanup(ids)
}

func (mgr TestTenantsManager) CleanupTenant(id string) {
	mgr.cleanup([]string{id})
}

func (mgr TestTenantsManager) GetIDByName(t require.TestingT, name string) string {
	val, ok := mgr.tenantsByName[name]
	require.True(t, ok)
	return val.ExternalTenant
}

func (mgr TestTenantsManager) GetDefaultTenantID() string {
	return mgr.tenantsByName[testDefaultTenant].ExternalTenant
}

func (mgr TestTenantsManager) GetSystemFetcherTenantID() string {
	return mgr.tenantsByName[TestSystemFetcherTenant].ExternalTenant
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

func (mgr TestTenantsManager) cleanup(ids []string) {
	dbCfg := persistence.DatabaseConfig{}
	if err := envconfig.Init(&dbCfg); err != nil {
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

	executeCleanupQuery(tx, deleteLabelDefinitionsQuery, ids)
	executeCleanupQuery(tx, deleteFormationsQuery, ids)

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func executeCleanupQuery(tx persistence.PersistenceTx, query string, ids []string) {
	q, args, err := sqlx.In(query, ids)
	if err != nil {
		log.Fatal(err)
	}

	q = sqlx.Rebind(sqlx.BindType("postgres"), q)

	// A tenant is considered initialized if there is any labelDefinitions associated with it.
	// On first request for a given tenant a labelDefinition for key scenario and value DEFAULT is created.
	// Therefore, once accessed a tenant is considered initialized. That's the reason we clean up (uninitialize) all the tests tenants here.
	// There is a test relying on this (testing tenants graphql query).
	if _, err = tx.ExecContext(context.TODO(), q, args...); err != nil {
		log.Fatal(err)
	}
}
