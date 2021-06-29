package product_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

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
