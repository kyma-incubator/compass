package label_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToEntity(t *testing.T) {
	stringValue := "foo"
	marshalledStringValue, err := json.Marshal(stringValue)
	require.NoError(t, err)

	arrayValue := []interface{}{"foo", "bar"}
	marshalledArrayValue, err := json.Marshal(arrayValue)
	require.NoError(t, err)

	version := 0

	// GIVEN
	testCases := []struct {
		Name               string
		Input              *model.Label
		Expected           *label.Entity
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given",
			Input:              fixLabelModel("1", arrayValue, version),
			Expected:           fixLabelEntity("1", marshalledArrayValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name:               "All properties given for formation template",
			Input:              fixFormationTemplateLabelModel(labelID, arrayValue, version),
			Expected:           fixFormationTemplateLabelEntity(labelID, marshalledArrayValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name:               "String value",
			Input:              fixLabelModel("1", stringValue, version),
			Expected:           fixLabelEntity("1", marshalledStringValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name: "Empty value",
			Input: &model.Label{
				ID:      "2",
				Key:     "foo",
				Version: version,
			},
			Expected:           nil,
			ExpectedErrMessage: "Value cannot be empty",
		},
		{
			Name: "Error",
			Input: &model.Label{
				Value: make(chan int),
			},
			Expected:           nil,
			ExpectedErrMessage: "while marshalling Value: json: unsupported type: chan int",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := label.NewConverter()

			// WHEN
			res, err := conv.ToEntity(testCase.Input)
			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
			}

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	stringValue := "foo"
	marshalledStringValue, err := json.Marshal(stringValue)
	require.NoError(t, err)

	arrayValue := []interface{}{"foo", "bar"}
	marshalledArrayValue, err := json.Marshal(arrayValue)
	require.NoError(t, err)

	version := 0

	// GIVEN
	testCases := []struct {
		Name               string
		Input              *label.Entity
		Expected           *model.Label
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given",
			Input:              fixLabelEntity("1", marshalledArrayValue, version),
			Expected:           fixLabelModel("1", arrayValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name:               "All properties given for formation template",
			Input:              fixFormationTemplateLabelEntity(labelID, marshalledArrayValue, version),
			Expected:           fixFormationTemplateLabelModel(labelID, arrayValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name:               "String value",
			Input:              fixLabelEntity("1", marshalledStringValue, version),
			Expected:           fixLabelModel("1", stringValue, version),
			ExpectedErrMessage: "",
		},
		{
			Name: "Nil value",
			Input: &label.Entity{
				ID:      "2",
				Key:     "foo",
				Version: version,
			},
			Expected:           nil,
			ExpectedErrMessage: "Value cannot be empty",
		},
		{
			Name: "Empty value",
			Input: &label.Entity{
				ID:      "2",
				Key:     "foo",
				Value:   "",
				Version: version,
			},
			Expected:           nil,
			ExpectedErrMessage: "Value cannot be empty",
		},
		{
			Name:               "Error",
			Input:              fixLabelEntity("1", []byte("{json"), version),
			Expected:           nil,
			ExpectedErrMessage: "while unmarshalling Value: invalid character 'j' looking for beginning of object key string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := label.NewConverter()

			// WHEN
			res, err := conv.FromEntity(testCase.Input)
			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
			}

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func fixLabelEntity(id string, value []byte, version int) *label.Entity {
	return &label.Entity{
		ID:    id,
		AppID: sql.NullString{},
		RuntimeID: sql.NullString{
			String: "321",
			Valid:  true,
		},
		TenantID: sql.NullString{},
		Key:      "test",
		Value:    string(value),
		Version:  version,
	}
}

func fixFormationTemplateLabelEntity(labelID string, value []byte, version int) *label.Entity {
	return &label.Entity{
		ID: labelID,
		FormationTemplateID: sql.NullString{
			String: testFormationTemplateID,
			Valid:  true,
		},
		Key:     key,
		Value:   string(value),
		Version: version,
	}
}

func fixLabelModel(id string, value interface{}, version int) *model.Label {
	return &model.Label{
		ID:         id,
		Key:        "test",
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   "321",
		Value:      value,
		Version:    version,
	}
}

func fixFormationTemplateLabelModel(labelID string, value interface{}, version int) *model.Label {
	return &model.Label{
		ID:         labelID,
		Key:        key,
		ObjectType: model.FormationTemplateLabelableObject,
		ObjectID:   testFormationTemplateID,
		Value:      value,
		Version:    version,
	}
}
