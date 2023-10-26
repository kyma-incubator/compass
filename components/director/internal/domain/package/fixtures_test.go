package ordpackage_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	packageID        = "packageID"
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	resourceHash     = "123456"
)

var (
	appID                = "appID"
	appTemplateVersionID = "appTemplateVersionID"
)

func fixEntityPackageForApp() *ordpackage.Entity {
	return fixEntityPackageWithTitleForApp("title")
}

func fixEntityPackageForAppTemplateVersion() *ordpackage.Entity {
	return fixEntityPackageWithTitleForAppTemplateVersion("title")
}

func fixEntityPackageWithTitleForApp(title string) *ordpackage.Entity {
	entity := fixEntityPackageWithTitle(title)
	entity.ApplicationID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityPackageWithTitleForAppTemplateVersion(title string) *ordpackage.Entity {
	entity := fixEntityPackageWithTitle(title)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return entity
}

func fixEntityPackageWithTitle(title string) *ordpackage.Entity {
	return &ordpackage.Entity{
		ID:                  packageID,
		OrdID:               ordID,
		Vendor:              sql.NullString{String: "vendorID", Valid: true},
		Title:               title,
		ShortDescription:    "short desc",
		Description:         "desc",
		Version:             "v1.0.5",
		PackageLinks:        repo.NewValidNullableString("{}"),
		Links:               repo.NewValidNullableString("[]"),
		LicenseType:         sql.NullString{String: "test", Valid: true},
		SupportInfo:         sql.NullString{String: "support-info", Valid: true},
		Tags:                repo.NewValidNullableString("[]"),
		Countries:           repo.NewValidNullableString("[]"),
		Labels:              repo.NewValidNullableString("{}"),
		PolicyLevel:         sql.NullString{String: "test", Valid: true},
		CustomPolicyLevel:   sql.NullString{},
		PartOfProducts:      repo.NewValidNullableString("[\"test\"]"),
		LineOfBusiness:      repo.NewValidNullableString("[]"),
		Industry:            repo.NewValidNullableString("[]"),
		ResourceHash:        repo.NewValidNullableString(resourceHash),
		DocumentationLabels: repo.NewValidNullableString("[]"),
	}
}

func fixNilModelPackage() *model.Package {
	return nil
}

func fixPackageModelForApp() *model.Package {
	return fixPackageModelWithTitleForApp("title")
}

func fixPackageModelForAppTemplateVersion() *model.Package {
	return fixPackageModelWithTitleForAppTemplateVersion("title")
}

func fixPackageModelWithTitleForAppTemplateVersion(title string) *model.Package {
	pkg := fixPackageModelWithTitle(title)
	pkg.ApplicationTemplateVersionID = &appTemplateVersionID
	return pkg
}
func fixPackageModelWithTitleForApp(title string) *model.Package {
	pkg := fixPackageModelWithTitle(title)
	pkg.ApplicationID = &appID
	return pkg
}

func fixPackageModelWithTitle(title string) *model.Package {
	vendorID := "vendorID"
	licenceType := "test"
	supportInfo := "support-info"
	policyLevel := "test"
	return &model.Package{
		ID:                  packageID,
		OrdID:               ordID,
		Vendor:              &vendorID,
		Title:               title,
		ShortDescription:    "short desc",
		Description:         "desc",
		Version:             "v1.0.5",
		PackageLinks:        json.RawMessage("{}"),
		Links:               json.RawMessage("[]"),
		LicenseType:         &licenceType,
		SupportInfo:         &supportInfo,
		Tags:                json.RawMessage("[]"),
		Countries:           json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		PolicyLevel:         &policyLevel,
		CustomPolicyLevel:   nil,
		PartOfProducts:      json.RawMessage("[\"test\"]"),
		LineOfBusiness:      json.RawMessage("[]"),
		Industry:            json.RawMessage("[]"),
		ResourceHash:        str.Ptr(resourceHash),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixPackageModelInput() *model.PackageInput {
	vendorID := "vendorID"
	licenceType := "test"
	supportInfo := "support-info"
	policyLevel := "test"
	return &model.PackageInput{
		OrdID:               ordID,
		Vendor:              &vendorID,
		Title:               "title",
		ShortDescription:    "short desc",
		Description:         "desc",
		Version:             "v1.0.5",
		PackageLinks:        json.RawMessage("{}"),
		Links:               json.RawMessage("[]"),
		LicenseType:         &licenceType,
		SupportInfo:         &supportInfo,
		Tags:                json.RawMessage("[]"),
		Countries:           json.RawMessage("[]"),
		Labels:              json.RawMessage("{}"),
		PolicyLevel:         &policyLevel,
		CustomPolicyLevel:   nil,
		PartOfProducts:      json.RawMessage("[\"test\"]"),
		LineOfBusiness:      json.RawMessage("[]"),
		Industry:            json.RawMessage("[]"),
		DocumentationLabels: json.RawMessage("[]"),
	}
}

func fixPackageColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "ord_id", "vendor", "title", "short_description",
		"description", "version", "package_links", "links", "licence_type", "tags", "countries", "labels", "policy_level",
		"custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash", "documentation_labels", "support_info"}
}

func fixPackageRowForApp() []driver.Value {
	return fixPackageRowWithTitleForApp("title")
}

func fixPackageRowForAppTemplateVersion() []driver.Value {
	return fixPackageRowWithTitleForAppTemplateVersion("title")
}

func fixPackageRowWithTitleForApp(title string) []driver.Value {
	return []driver.Value{packageID, appID, repo.NewValidNullableString(""), ordID, "vendorID", title, "short desc", "desc", "v1.0.5",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString("[]"), "test", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}"),
		"test", nil, repo.NewValidNullableString("[\"test\"]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]"), "support-info"}
}

func fixPackageRowWithTitleForAppTemplateVersion(title string) []driver.Value {
	return []driver.Value{packageID, repo.NewValidNullableString(""), appTemplateVersionID, ordID, "vendorID", title, "short desc", "desc", "v1.0.5",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString("[]"), "test", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}"),
		"test", nil, repo.NewValidNullableString("[\"test\"]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]"), "support-info"}
}

func fixPackageUpdateArgs() []driver.Value {
	return []driver.Value{"vendorID", "title", "short desc", "desc", "v1.0.5", repo.NewValidNullableString("{}"), repo.NewValidNullableString("[]"),
		"test", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}"), "test", nil, repo.NewValidNullableString("[\"test\"]"),
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]"), "support-info"}
}
