package tenant

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/vrischmann/envconfig"

	"github.com/stretchr/testify/require"
)

const (
	testProvider           = "Compass Tests"
	testDefaultTenant      = "Test Default"
	deleteLabelDefinitions = `DELETE FROM public.label_definitions WHERE tenant_id IN (%s);`

	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"

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
	Status         TenantStatus `db:"status"`
}

type TenantStatus string

var TestTenants TestTenantsManager

type TestTenantsManager struct {
	tenantsByName map[string]Tenant
}

func (mgr *TestTenantsManager) Init() {
	mgr.Cleanup()
	mgr.tenantsByName = map[string]Tenant{
		testDefaultTenant: {
			ID:             "5577cf46-4f78-45fa-b55f-a42a3bdba868",
			Name:           testDefaultTenant,
			ExternalTenant: "5577cf46-4f78-45fa-b55f-a42a3bdba868",
			ProviderName:   testProvider,
			Status:         Active,
		},
		TenantSeparationTenantName: {
			ID:             "f1c4b5be-b0e1-41f9-b0bc-b378200dcca0",
			Name:           TenantSeparationTenantName,
			ExternalTenant: "f1c4b5be-b0e1-41f9-b0bc-b378200dcca0",
			ProviderName:   testProvider,
			Status:         Active,
		},
		ApplicationsForRuntimeTenantName: {
			ID:             "5984a414-1eed-4972-af2c-b2b6a415c7d7",
			Name:           ApplicationsForRuntimeTenantName,
			ExternalTenant: "5984a414-1eed-4972-af2c-b2b6a415c7d7",
			ProviderName:   testProvider,
			Status:         Active,
		},
		ListLabelDefinitionsTenantName: {
			ID:             "3f641cf5-2d14-4e0f-a122-16e7569926f1",
			Name:           ListLabelDefinitionsTenantName,
			ExternalTenant: "3f641cf5-2d14-4e0f-a122-16e7569926f1",
			ProviderName:   testProvider,
			Status:         Active,
		},
		DeleteLastScenarioForApplicationTenantName: {
			ID:             "0403be1e-f854-475e-9074-922120277af5",
			Name:           DeleteLastScenarioForApplicationTenantName,
			ExternalTenant: "0403be1e-f854-475e-9074-922120277af5",
			ProviderName:   testProvider,
			Status:         Active,
		},
		DeleteAutomaticScenarioAssignmentForScenarioTenantName: {
			ID:             "d08e4cb6-a77f-4a07-b021-e3317a373597",
			Name:           DeleteAutomaticScenarioAssignmentForScenarioTenantName,
			ExternalTenant: "d08e4cb6-a77f-4a07-b021-e3317a373597",
			ProviderName:   testProvider,
			Status:         Active,
		},
		DeleteAutomaticScenarioAssignmentForSelectorTenantName: {
			ID:             "d9553135-6115-4c67-b4d9-962c00f3725f",
			Name:           DeleteAutomaticScenarioAssignmentForSelectorTenantName,
			ExternalTenant: "d9553135-6115-4c67-b4d9-962c00f3725f",
			ProviderName:   testProvider,
			Status:         Active,
		},
		AutomaticScenarioAssigmentForRuntimeTenantName: {
			ID:             "8c733a45-d988-4472-af10-1256b82c70c0",
			Name:           AutomaticScenarioAssigmentForRuntimeTenantName,
			ExternalTenant: "8c733a45-d988-4472-af10-1256b82c70c0",
			ProviderName:   testProvider,
			Status:         Active,
		},
		AutomaticScenarioAssignmentQueriesTenantName: {
			ID:             "8263cc13-5698-4a2d-9257-e8e76b543e88",
			Name:           AutomaticScenarioAssignmentQueriesTenantName,
			ExternalTenant: "8263cc13-5698-4a2d-9257-e8e76b543e88",
			ProviderName:   testProvider,
			Status:         Active,
		},
		GetScenariosLabelDefinitionCreatesOneIfNotExistsTenantName: {
			ID:             "2263cc13-5698-4a2d-9257-e8e76b543e33",
			Name:           GetScenariosLabelDefinitionCreatesOneIfNotExistsTenantName,
			ExternalTenant: "2263cc13-5698-4a2d-9257-e8e76b543e33",
			ProviderName:   testProvider,
			Status:         Active,
		},
		AutomaticScenarioAssignmentsWholeScenarioTenantName: {
			ID:             "65a63692-c00a-4a7d-8376-8615ee37f45c",
			Name:           AutomaticScenarioAssignmentsWholeScenarioTenantName,
			ExternalTenant: "65a63692-c00a-4a7d-8376-8615ee37f45c",
			ProviderName:   testProvider,
			Status:         Active,
		},
		ApplicationsForRuntimeWithHiddenAppsTenantName: {
			ID:             "7e1f2df8-36dc-4e40-8be3-d1555d50c91c",
			Name:           ApplicationsForRuntimeWithHiddenAppsTenantName,
			ExternalTenant: "7e1f2df8-36dc-4e40-8be3-d1555d50c91c",
			ProviderName:   testProvider,
			Status:         Active,
		},
		TenantsQueryInitializedTenantName: {
			ID:             "8cf0c909-f816-4fe3-a507-a7917ccd8380",
			Name:           TenantsQueryInitializedTenantName,
			ExternalTenant: "8cf0c909-f816-4fe3-a507-a7917ccd8380",
			ProviderName:   testProvider,
			Status:         Active,
		},
		TenantsQueryNotInitializedTenantName: {
			ID:             "72329135-27fd-4284-9bcb-37ea8d6307d0",
			Name:           TenantsQueryNotInitializedTenantName,
			ExternalTenant: "72329135-27fd-4284-9bcb-37ea8d6307d0",
			ProviderName:   testProvider,
			Status:         Active,
		},
		TestDeleteApplicationIfInScenario: {
			ID:             "0d597250-6b2d-4d89-9c54-e23cb497cd01",
			Name:           TestDeleteApplicationIfInScenario,
			ExternalTenant: "0d597250-6b2d-4d89-9c54-e23cb497cd01",
			ProviderName:   testProvider,
			Status:         Active,
		},
	}
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
	args := make([]interface{}, 0, len(tenants))
	var query string
	for i, tnt := range tenants {
		args = append(args, tnt.ID)
		query += fmt.Sprintf("$%d, ", i+1)
	}
	query = strings.TrimSuffix(query, ", ")

	// A tenant is considered initialized if there is any labelDefinitions associated with it.
	// On first request for a given tenant a labelDefinition for key scenario and value DEFAULT is created.
	// Therefore once accessed a tenant is considered initialized. That's the reason we clean up (uninitialize) all the tests tenants here.
	// There is a test relying on this (testing tenants graphql query).
	_, err = tx.ExecContext(context.TODO(), fmt.Sprintf(deleteLabelDefinitions, query), args...)
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
	return val.ID
}

func (mgr TestTenantsManager) GetDefaultTenantID() string {
	return mgr.tenantsByName[testDefaultTenant].ID
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
