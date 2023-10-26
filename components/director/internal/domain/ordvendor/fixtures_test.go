package ordvendor_test

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	vendorID             = "vendorID"
	tenantID             = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appID                = "appID"
	appTemplateVersionID = "appTemplateVersionID"
	ordID                = "com.compass.v1"
	externalTenantID     = "externalTenantID"
	partners             = `["microsoft:vendor:Microsoft:"]`
)

func fixEntityVendor() *ordvendor.Entity {
	return &ordvendor.Entity{
		ID:                  vendorID,
		OrdID:               ordID,
		ApplicationID:       repo.NewValidNullableString(appID),
		Title:               "title",
		Partners:            repo.NewValidNullableString(partners),
		Tags:                repo.NewValidNullableString("[]"),
		Labels:              repo.NewValidNullableString("{}"),
		DocumentationLabels: repo.NewValidNullableString("[]"),
	}
}

func fixEntityVendorWithTitle(title string) *ordvendor.Entity {
	return &ordvendor.Entity{
		ID:                  vendorID,
		OrdID:               ordID,
		Title:               title,
		Partners:            repo.NewValidNullableString(partners),
		Tags:                repo.NewValidNullableString("[]"),
		Labels:              repo.NewValidNullableString("{}"),
		DocumentationLabels: repo.NewValidNullableString("[]"),
	}
}

func fixEntityVendorWithTitleForApp(title string) *ordvendor.Entity {
	entity := fixEntityVendorWithTitle(title)
	entity.ApplicationID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityVendorWithTitleForAppTemplateVersion(title string) *ordvendor.Entity {
	entity := fixEntityVendorWithTitle(title)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return entity
}

func fixVendorModelWithTitle(title string) *model.Vendor {
	return &model.Vendor{
		ID:                  vendorID,
		OrdID:               ordID,
		Title:               title,
		Partners:            json.RawMessage(partners),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixVendorModelWithTitleForAppTemplateVersion(title string) *model.Vendor {
	vendor := fixVendorModelWithTitle(title)
	vendor.ApplicationTemplateVersionID = str.Ptr(appTemplateVersionID)
	return vendor
}

func fixVendorModelWithTitleForApp(title string) *model.Vendor {
	vendor := fixVendorModelWithTitle(title)
	vendor.ApplicationID = str.Ptr(appID)
	return vendor
}

func fixVendorModel() *model.Vendor {
	return &model.Vendor{
		ID:                  vendorID,
		OrdID:               ordID,
		Title:               "title",
		Partners:            json.RawMessage(partners),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixVendorModelForApp() *model.Vendor {
	vendor := fixVendorModel()
	vendor.ApplicationID = str.Ptr(appID)
	return vendor
}

func fixVendorModelForAppTemplateVersion() *model.Vendor {
	vendor := fixVendorModel()
	vendor.ApplicationTemplateVersionID = str.Ptr(appTemplateVersionID)
	return vendor
}

func fixGlobalVendorModel() *model.Vendor {
	return &model.Vendor{
		ID:                  vendorID,
		OrdID:               ordID,
		Title:               "title",
		Partners:            json.RawMessage(partners),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixNilModelVendor() *model.Vendor {
	return nil
}

func fixVendorModelInput() *model.VendorInput {
	return &model.VendorInput{
		OrdID:               ordID,
		Title:               "title",
		Partners:            json.RawMessage(partners),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixVendorColumns() []string {
	return []string{"ord_id", "app_id", "app_template_version_id", "title", "labels", "partners", "id", "tags", "documentation_labels"}
}

func fixVendorRow() []driver.Value {
	return []driver.Value{ordID, appID, repo.NewValidNullableString(""), "title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), vendorID, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixVendorRowWithTitleForApp(title string) []driver.Value {
	return []driver.Value{ordID, appID, repo.NewValidNullableString(""), title, repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), vendorID, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixVendorRowWithTitleForAppTemplateVersion(title string) []driver.Value {
	return []driver.Value{ordID, repo.NewValidNullableString(""), appTemplateVersionID, title, repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), vendorID, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixVendorUpdateArgs() []driver.Value {
	return []driver.Value{"title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}
