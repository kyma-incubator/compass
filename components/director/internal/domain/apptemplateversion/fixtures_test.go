package apptemplateversion_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

const (
	appTemplateVersionID = "44444444-1111-2222-3333-51d5356e7e09"
	testTitle            = "testTitle"
	appTemplateID        = "58963c6f-24f6-4128-a05c-51d5356e7e09"
	testVersion          = "2306"
)

var (
	mockedTimestamp    = time.Now().String()
	testCorrelationIDs = json.RawMessage(`["one"]`)
	testError          = errors.New("test error")
	testTableColumns   = []string{"id", "version", "title", "correlation_ids", "release_date", "created_at", "app_template_id"}
)

func fixModelApplicationTemplateVersion(id string) *model.ApplicationTemplateVersion {
	return &model.ApplicationTemplateVersion{
		ID:                    id,
		Version:               testVersion,
		Title:                 str.Ptr(testTitle),
		ReleaseDate:           &mockedTimestamp,
		CorrelationIDs:        testCorrelationIDs,
		CreatedAt:             mockedTimestamp,
		ApplicationTemplateID: appTemplateID,
	}
}

func fixModelApplicationTemplateVersionInput() *model.ApplicationTemplateVersionInput {
	return &model.ApplicationTemplateVersionInput{
		Version:        testVersion,
		Title:          str.Ptr(testTitle),
		ReleaseDate:    &mockedTimestamp,
		CorrelationIDs: testCorrelationIDs,
	}
}

func fixEntityApplicationTemplateVersion(t *testing.T, id string) *apptemplateversion.Entity {
	marshalledCorrelationIDs, err := json.Marshal(testCorrelationIDs)
	require.NoError(t, err)

	return &apptemplateversion.Entity{
		ID:                    id,
		Version:               testVersion,
		Title:                 repo.NewNullableString(str.Ptr(testTitle)),
		ReleaseDate:           repo.NewValidNullableString(mockedTimestamp),
		CorrelationIDs:        repo.NewValidNullableString(string(marshalledCorrelationIDs)),
		CreatedAt:             mockedTimestamp,
		ApplicationTemplateID: appTemplateID,
	}
}

func fixColumns() []string {
	return []string{"id", "version", "title", "correlation_ids", "release_date", "created_at", "app_template_id"}
}

func fixAppTemplateVersionCreateArgs(entity apptemplateversion.Entity) []driver.Value {
	return []driver.Value{entity.ID, entity.Version, entity.Title, entity.CorrelationIDs, entity.ReleaseDate, entity.CreatedAt, entity.ApplicationTemplateID}
}

func fixSQLRows(entities []apptemplateversion.Entity) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, entity := range entities {
		out.AddRow(entity.ID, entity.Version, entity.Title, entity.CorrelationIDs, entity.ReleaseDate, entity.CreatedAt, entity.ApplicationTemplateID)
	}
	return out
}
