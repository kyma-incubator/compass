package product_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	productID            = "productID"
	tenantID             = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appID                = "appID"
	appTemplateVersionID = "appTemplateVersionID"
	ordID                = "com.compass.v1"
	externalTenantID     = "externalTenantID"
	correlationIDs       = `["id1", "id2"]`
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
		Title:            title,
		ShortDescription: "short desc",
		Description:      "desc",
		Vendor:           "vendorID",
		Parent: sql.NullString{
			String: "parent",
			Valid:  true,
		},
		CorrelationIDs: sql.NullString{
			String: correlationIDs,
			Valid:  true,
		},
		Tags:                repo.NewValidNullableString("[]"),
		Labels:              repo.NewValidNullableString("{}"),
		DocumentationLabels: repo.NewValidNullableString("{}"),
	}
}

func fixEntityProductForApp() *product.Entity {
	entity := fixEntityProduct()
	entity.ApplicationID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityProductWithTitleForApp(title string) *product.Entity {
	entity := fixEntityProductWithTitle(title)
	entity.ApplicationID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityProductWithTitleForAppTemplateVersion(title string) *product.Entity {
	entity := fixEntityProductWithTitle(title)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return entity
}

func fixProductModelForApp() *model.Product {
	return fixProductModelWithTitleForApp("title")
}

func fixProductModelForAppTemplateVersion() *model.Product {
	return fixProductModelWithTitleForAppTemplateVersion("title")
}

func fixProductModelWithTitle(title string) *model.Product {
	parent := "parent"
	return &model.Product{
		ID:                  productID,
		OrdID:               ordID,
		Title:               title,
		ShortDescription:    "short desc",
		Description:         "desc",
		Vendor:              "vendorID",
		Parent:              &parent,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("{}"),
	}
}

func fixProductModelWithTitleForApp(title string) *model.Product {
	product := fixProductModelWithTitle(title)
	product.ApplicationID = str.Ptr(appID)
	return product
}

func fixProductModelWithTitleForAppTemplateVersion(title string) *model.Product {
	product := fixProductModelWithTitle(title)
	product.ApplicationTemplateVersionID = str.Ptr(appTemplateVersionID)
	return product
}

func fixGlobalProductModel() *model.Product {
	parent := "parent"
	return &model.Product{
		ID:                  productID,
		OrdID:               ordID,
		Title:               "title",
		ShortDescription:    "short desc",
		Description:         "desc",
		Vendor:              "vendorID",
		Parent:              &parent,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("{}"),
	}
}

func fixProductModelInput() *model.ProductInput {
	parent := "parent"
	return &model.ProductInput{
		OrdID:               ordID,
		Title:               "title",
		ShortDescription:    "short desc",
		Description:         "desc",
		Vendor:              "vendorID",
		Parent:              &parent,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("{}"),
	}
}

func fixProductColumns() []string {
	return []string{"ord_id", "app_id", "app_template_version_id", "title", "short_description", "description", "vendor", "parent", "labels", "correlation_ids", "id", "tags", "documentation_labels"}
}

func fixProductRow() []driver.Value {
	return fixProductRowWithTitleForApp("title")
}

func fixProductRowWithTitleForApp(title string) []driver.Value {
	return []driver.Value{ordID, appID, repo.NewValidNullableString(""), title, "short desc", "desc", "vendorID", "parent",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIDs), productID, repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}")}
}

func fixProductRowWithTitleForAppTemplateVersion(title string) []driver.Value {
	return []driver.Value{ordID, repo.NewValidNullableString(""), appTemplateVersionID, title, "short desc", "desc", "vendorID", "parent",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIDs), productID, repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}")}
}

func fixProductUpdateArgs() []driver.Value {
	return []driver.Value{"title", "short desc", "desc", "vendorID", "parent", repo.NewValidNullableString("{}"), repo.NewValidNullableString(correlationIDs), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}")}
}
