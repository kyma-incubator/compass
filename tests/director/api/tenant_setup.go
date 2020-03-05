package api

import (
	"log"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const (
	insertQuery  = `INSERT INTO public.business_tenant_mappings (id, external_name, external_tenant, provider_name, status) VALUES ($1, $2, $3, $4, $5)`
	deleteQuery  = `DELETE FROM public.business_tenant_mappings WHERE provider_name = $1`
	testProvider = "Compass Tests"
)

type Tenant struct {
	ID             string       `db:"id"`
	Name           string       `db:"external_name"`
	ExternalTenant string       `db:"external_tenant"`
	ProviderName   string       `db:"provider_name"`
	Status         TenantStatus `db:"status"`
}

type TenantStatus string

const (
	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"
)

func insertTenants(transact persistence.Transactioner) {
	testTenants := fixTestTenants()

	tx, err := transact.Begin()
	for _, v := range testTenants {
		_, err := tx.Exec(insertQuery, v.ID, v.Name, v.ExternalTenant, v.ProviderName, v.Status)

		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

}

func deleteTenants(transact persistence.Transactioner) {

	tx, err := transact.Begin()

	_, err = tx.Exec(deleteQuery, testProvider)

	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()

}

func fixTestTenants() []Tenant {
	return []Tenant{
		{
			ID:             "f1c4b5be-b0e1-41f9-b0bc-b378200dcca0",
			Name:           "Test1",
			ExternalTenant: "f1c4b5be-b0e1-41f9-b0bc-b378200dcca0",
			ProviderName:   testProvider,
			Status:         Active,
		},
		{
			ID:             "5984a414-1eed-4972-af2c-b2b6a415c7d7",
			Name:           "Test2",
			ExternalTenant: "5984a414-1eed-4972-af2c-b2b6a415c7d7",
			ProviderName:   testProvider,
			Status:         Active,
		},
		{
			ID:             "2bf03de1-23b1-4063-9d3e-67096800accc",
			Name:           "Test3",
			ExternalTenant: "2bf03de1-23b1-4063-9d3e-67096800accc",
			ProviderName:   testProvider,
			Status:         Active,
		},
		{
			ID:             "f739b36c-813f-4fc3-996e-dd03c7d13aa0",
			Name:           "Test4",
			ExternalTenant: "f739b36c-813f-4fc3-996e-dd03c7d13aa0",
			ProviderName:   testProvider,
			Status:         Active,
		},
	}
}

func tenantsToGraphql(tenants []Tenant) []*graphql.Tenant {
	var toReturn []*graphql.Tenant

	for k, _ := range tenants {
		toReturn = append(toReturn, &graphql.Tenant{ID: tenants[k].ID, Name: &tenants[k].Name})
	}

	return toReturn
}
