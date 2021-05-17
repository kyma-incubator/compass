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

var boolPtr = true

func fixEntityVendor() *ordvendor.Entity {
	return &ordvendor.Entity{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		SapPartner:    repo.NewNullableBool(&boolPtr),
		Labels:        repo.NewValidNullableString("{}"),
	}
}

func fixVendorModel() *model.Vendor {
	return &model.Vendor{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         "title",
		SapPartner:    &boolPtr,
		Labels:        json.RawMessage("{}"),
	}
}

func fixVendorModelInput() *model.VendorInput {
	return &model.VendorInput{
		OrdID:      ordID,
		Title:      "title",
		SapPartner: &boolPtr,
		Labels:     json.RawMessage("{}"),
	}
}

func fixVendorColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "title", "labels", "sap_partner"}
}

func fixVendorRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "title", repo.NewValidNullableString("{}"), repo.NewNullableBool(&boolPtr)}
}

func fixVendorUpdateArgs() []driver.Value {
	return []driver.Value{"title", repo.NewValidNullableString("{}"), repo.NewNullableBool(&boolPtr)}
}
