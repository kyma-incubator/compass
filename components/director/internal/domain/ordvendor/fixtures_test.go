package ordvendor_test

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
)

func fixEntityVendor() *ordvendor.Entity {
	return &ordvendor.Entity{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		Type:          "type",
		Labels:        repo.NewValidNullableString("{}"),
	}
}

func fixVendorModel() *model.Vendor {
	return &model.Vendor{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		Type:          "type",
		Labels:        json.RawMessage("{}"),
	}
}

func fixVendorModelInput() *model.VendorInput {
	return &model.VendorInput{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		Type:          "type",
		Labels:        json.RawMessage("{}"),
	}
}

func fixVendorColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "title", "type", "labels"}
}

func fixVendorRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "title", "type", repo.NewValidNullableString("{}")}
}

func fixVendorUpdateArgs() []driver.Value {
	return []driver.Value{"title", "type", repo.NewValidNullableString("{}")}
}
