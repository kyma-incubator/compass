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
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
)

func fixEntityProduct() *product.Entity {
	return &product.Entity{
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
		PPMSObjectID: sql.NullString{
			String: "ppms_id",
			Valid:  true,
		},
		Labels: repo.NewValidNullableString("{}"),
	}
}

func fixProductModel() *model.Product {
	parent := "parent"
	ppmsID := "ppms_id"
	return &model.Product{
		OrdID:            ordID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            "title",
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent:           &parent,
		PPMSObjectID:     &ppmsID,
		Labels:           json.RawMessage("{}"),
	}
}

func fixProductModelInput() *model.ProductInput {
	parent := "parent"
	ppmsID := "ppms_id"
	return &model.ProductInput{
		OrdID:            ordID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            "title",
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent:           &parent,
		PPMSObjectID:     &ppmsID,
		Labels:           json.RawMessage("{}"),
	}
}

func fixProductColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "title", "short_description", "vendor", "parent", "sap_ppms_object_id", "labels"}
}

func fixProductRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "title", "short desc", "vendorID", "parent", "ppms_id",
		repo.NewValidNullableString("{}")}
}

func fixProductUpdateArgs() []driver.Value {
	return []driver.Value{"title", "short desc", "vendorID", "parent", "ppms_id", repo.NewValidNullableString("{}")}
}
