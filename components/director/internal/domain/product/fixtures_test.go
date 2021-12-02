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
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	correlationIDs   = `["id1", "id2"]`
)

func fixNilModelProduct() *model.Product {
	return nil
}

func fixEntityProduct() *product.Entity {
	return fixEntityProductWithTitle("title")
}

func fixEntityProductWithTitle(title string) *product.Entity {
	return &product.Entity{
		ID:               productID,
		OrdID:            ordID,
		ApplicationID:    appID,
		Title:            title,
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent: sql.NullString{
			String: "parent",
			Valid:  true,
		},
		CorrelationIDs: sql.NullString{
			String: correlationIDs,
			Valid:  true,
		},
		Labels: repo.NewValidNullableString("{}"),
	}
}

func fixProductModel() *model.Product {
	return fixProductModelWithTitle("title")
}

func fixProductModelWithTitle(title string) *model.Product {
	parent := "parent"
	return &model.Product{
		ID:               productID,
		OrdID:            ordID,
		ApplicationID:    appID,
		Title:            title,
		ShortDescription: "short desc",
		Vendor:           "vendorID",
		Parent:           &parent,
		CorrelationIDs:   json.RawMessage(correlationIDs),
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
		CorrelationIDs:   json.RawMessage(correlationIDs),
		Labels:           json.RawMessage("{}"),
	}
}

func fixProductColumns() []string {
	return []string{"ord_id", "app_id", "title", "short_description", "vendor", "parent", "labels", "correlation_ids", "id"}
}

func fixProductRow() []driver.Value {
	return fixProductRowWithTitle("title")
}

func fixProductRowWithTitle(title string) []driver.Value {
	return []driver.Value{ordID, appID, title, "short desc", "vendorID", "parent",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIDs), productID}
}

func fixProductUpdateArgs() []driver.Value {
	return []driver.Value{"title", "short desc", "vendorID", "parent", repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIDs)}
}
