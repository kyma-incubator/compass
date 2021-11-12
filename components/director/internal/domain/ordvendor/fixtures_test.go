package ordvendor_test

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	vendorID         = "vendorID"
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	partners         = `["microsoft:vendor:Microsoft:"]`
)

func fixEntityVendor() *ordvendor.Entity {
	return &ordvendor.Entity{
		ID:            vendorID,
		OrdID:         ordID,
		ApplicationID: appID,
		Title:         "title",
		Partners:      repo.NewValidNullableString(partners),
		Labels:        repo.NewValidNullableString("{}"),
	}
}

func fixVendorModel() *model.Vendor {
	return &model.Vendor{
		ID:            vendorID,
		OrdID:         ordID,
		ApplicationID: appID,
		Title:         "title",
		Partners:      json.RawMessage(partners),
		Labels:        json.RawMessage("{}"),
	}
}

func fixNilModelVendor() *model.Vendor {
	return nil
}

func fixVendorModelInput() *model.VendorInput {
	return &model.VendorInput{
		OrdID:    ordID,
		Title:    "title",
		Partners: json.RawMessage(partners),
		Labels:   json.RawMessage("{}"),
	}
}

func fixVendorColumns() []string {
	return []string{"ord_id", "app_id", "title", "labels", "partners", "id"}
}

func fixVendorRow() []driver.Value {
	return []driver.Value{ordID, appID, "title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners), vendorID}
}

func fixVendorUpdateArgs() []driver.Value {
	return []driver.Value{"title", repo.NewValidNullableString("{}"), repo.NewValidNullableString(partners)}
}
