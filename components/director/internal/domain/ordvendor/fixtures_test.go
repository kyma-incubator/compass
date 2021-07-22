package ordvendor_test

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	vendorID         = "vendorID"
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	partners         = `["microsoft:vendor:Microsoft:"]`
)

func fixEntityVendor() *ordvendor.Entity {
	return &ordvendor.Entity{
		ID:            vendorID,
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		Partners:      repo.NewValidNullableString(partners),
		Labels:        repo.NewValidNullableString("{}"),
	}
}

func fixVendorModel() *model.Vendor {
	return &model.Vendor{
		ID:            vendorID,
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		Partners:      json.RawMessage(partners),
		Labels:        json.RawMessage("{}"),
	}
}

func fixVendorModelInput() *model.VendorInput {
	return &model.VendorInput{
		OrdID:    ordID,
		Title:    "title",
		Partners: json.RawMessage(partners),
		Labels:   json.RawMessage("{}"),
	}
}

func fixVendorColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "title", "labels", "partners", "id"}
}

func fixVendorRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), vendorID}
}

func fixVendorUpdateArgs() []driver.Value {
	return []driver.Value{"title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners)}
}

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
