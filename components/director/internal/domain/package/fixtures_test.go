package ordpackage_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	packageID        = "packageID"
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	resourceHash     = "123456"
)

func fixEntityPackage() *ordpackage.Entity {
	return &ordpackage.Entity{
		ID:                packageID,
		TenantID:          tenantID,
		ApplicationID:     appID,
		OrdID:             ordID,
		Vendor:            sql.NullString{String: "vendorID", Valid: true},
		Title:             "title",
		ShortDescription:  "short desc",
		Description:       "desc",
		Version:           "v1.0.5",
		PackageLinks:      repo.NewValidNullableString("{}"),
		Links:             repo.NewValidNullableString("[]"),
		LicenseType:       sql.NullString{String: "test", Valid: true},
		Tags:              repo.NewValidNullableString("[]"),
		Countries:         repo.NewValidNullableString("[]"),
		Labels:            repo.NewValidNullableString("{}"),
		PolicyLevel:       "test",
		CustomPolicyLevel: sql.NullString{},
		PartOfProducts:    repo.NewValidNullableString("[\"test\"]"),
		LineOfBusiness:    repo.NewValidNullableString("[]"),
		Industry:          repo.NewValidNullableString("[]"),
		ResourceHash:      repo.NewValidNullableString(resourceHash),
	}
}

func fixPackageModel() *model.Package {
	vendorID := "vendorID"
	licenceType := "test"
	return &model.Package{
		ID:                packageID,
		TenantID:          tenantID,
		ApplicationID:     appID,
		OrdID:             ordID,
		Vendor:            &vendorID,
		Title:             "title",
		ShortDescription:  "short desc",
		Description:       "desc",
		Version:           "v1.0.5",
		PackageLinks:      json.RawMessage("{}"),
		Links:             json.RawMessage("[]"),
		LicenseType:       &licenceType,
		Tags:              json.RawMessage("[]"),
		Countries:         json.RawMessage("[]"),
		Labels:            json.RawMessage("{}"),
		PolicyLevel:       "test",
		CustomPolicyLevel: nil,
		PartOfProducts:    json.RawMessage("[\"test\"]"),
		LineOfBusiness:    json.RawMessage("[]"),
		Industry:          json.RawMessage("[]"),
		ResourceHash:      str.Ptr(resourceHash),
	}
}

func fixPackageModelInput() *model.PackageInput {
	vendorID := "vendorID"
	licenceType := "test"
	return &model.PackageInput{
		OrdID:             ordID,
		Vendor:            &vendorID,
		Title:             "title",
		ShortDescription:  "short desc",
		Description:       "desc",
		Version:           "v1.0.5",
		PackageLinks:      json.RawMessage("{}"),
		Links:             json.RawMessage("[]"),
		LicenseType:       &licenceType,
		Tags:              json.RawMessage("[]"),
		Countries:         json.RawMessage("[]"),
		Labels:            json.RawMessage("{}"),
		PolicyLevel:       "test",
		CustomPolicyLevel: nil,
		PartOfProducts:    json.RawMessage("[\"test\"]"),
		LineOfBusiness:    json.RawMessage("[]"),
		Industry:          json.RawMessage("[]"),
	}
}

func fixPackageColumns() []string {
	return []string{"id", "tenant_id", "app_id", "ord_id", "vendor", "title", "short_description",
		"description", "version", "package_links", "links", "licence_type", "tags", "countries", "labels", "policy_level",
		"custom_policy_level", "part_of_products", "line_of_business", "industry", "resource_hash"}
}

func fixPackageRow() []driver.Value {
	return []driver.Value{packageID, tenantID, appID, ordID, "vendorID", "title", "short desc", "desc", "v1.0.5",
		repo.NewValidNullableString("{}"), repo.NewValidNullableString("[]"), "test", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}"),
		"test", nil, repo.NewValidNullableString("[\"test\"]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash)}
}

func fixPackageUpdateArgs() []driver.Value {
	return []driver.Value{"vendorID", "title", "short desc", "desc", "v1.0.5", repo.NewValidNullableString("{}"), repo.NewValidNullableString("[]"),
		"test", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("{}"), "test", nil, repo.NewValidNullableString("[\"test\"]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString(resourceHash)}
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
