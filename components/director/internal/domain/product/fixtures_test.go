package product_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	productID        = "productID"
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	correlationIds   = `["id1", "id2"]`
)

func fixEntityProduct() *product.Entity {
	return &product.Entity{
		ID:               productID,
		OrdID:            ordID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            "title",
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent: sql.NullString{
			String: "parent",
			Valid:  true,
		},
		CorrelationIds: sql.NullString{
			String: correlationIds,
			Valid:  true,
		},
		Labels: repo.NewValidNullableString("{}"),
	}
}

func fixProductModel() *model.Product {
	parent := "parent"
	return &model.Product{
		ID:               productID,
		OrdID:            ordID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            "title",
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent:           &parent,
		CorrelationIds:   json.RawMessage(correlationIds),
		Labels:           json.RawMessage("{}"),
	}
}

func fixProductModelInput() *model.ProductInput {
	parent := "parent"
	return &model.ProductInput{
		OrdID:            ordID,
		Title:            "title",
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent:           &parent,
		CorrelationIds:   json.RawMessage(correlationIds),
		Labels:           json.RawMessage("{}"),
	}
}

func fixProductColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "title", "short_description", "vendor", "parent", "labels", "correlation_ids", "id"}
}

func fixProductRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "title", "short desc", "vendorID", "parent",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIds), productID}
}

func fixProductUpdateArgs() []driver.Value {
	return []driver.Value{"title", "short desc", "vendorID", "parent", repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIds)}
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
