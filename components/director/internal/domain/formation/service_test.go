package formation_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	TestFormationName = "test-formation"
)

func TestServiceCreateFormation(t *testing.T) {
	t.Run("success when no labeldef exists", func(t *testing.T) {
		// GIVEN
		//mockLabelDefRepository := &automock.LabelDefRepository{}
		//mockLabelDefService := &automock.LabelDefService{}
		////mockUID := &automock.UIDService{}
		//defer mockLabelDefRepository.AssertExpectations(t)
		//defer mockLabelDefService.AssertExpectations(t)
		////defer mockUID.AssertExpectations(t)
		//
		//in := model.Formation{
		//	Name: TestFormationName,
		//}
		//
		//tnt := "tenant"
		//errNotFound := errors.New("Not Found")
		//
		//expected := &model.Formation{
		//	Name: "test-formation",
		//}
		//ctx := context.TODO()
		//mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, errNotFound)
		//mockLabelDefService.On("CreateWithFormations", ctx, tnt, []string{TestFormationName}).Return(&in, nil)
		//sut := formation.NewService(nil, mockLabelDefRepository, nil, nil, mockLabelDefService)
		//// WHEN
		//actual, err := sut.CreateFormation(ctx, tnt, in)
		//// THEN
		//require.NoError(t, err)
		//assert.Equal(t, expected, actual)
	})
	t.Run("success when labeldef exists", func(t *testing.T) {})
	t.Run("error when labeldef is missing and can not create it", func(t *testing.T) {})
	t.Run("error when labeldef is missing and can not create it", func(t *testing.T) {})
	t.Run("error when can not get labeldef", func(t *testing.T) {})
	t.Run("error when labeldef's schema is missing", func(t *testing.T) {})
	t.Run("error when validating existing labels against the schema", func(t *testing.T) {})
	t.Run("error when validating automatic scenario assignment against the schema", func(t *testing.T) {})
	t.Run("error when version is already updated", func(t *testing.T) {})
}

func TestServiceDeleteFormation(t *testing.T) {
	t.Run("success when no labeldef exists", func(t *testing.T) {})
	t.Run("success when labeldef exists", func(t *testing.T) {})
	t.Run("error when labeldef is missing and can not create it", func(t *testing.T) {})
	t.Run("error when labeldef is missing and can not create it", func(t *testing.T) {})
	t.Run("error when can not get labeldef", func(t *testing.T) {})
	t.Run("error when labeldef's schema is missing", func(t *testing.T) {})
	t.Run("error when validating existing labels against the schema", func(t *testing.T) {})
	t.Run("error when validating automatic scenario assignment against the schema", func(t *testing.T) {})
	t.Run("error when version is already updated", func(t *testing.T) {})
}

func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixFormation() *model.Formation {
	return &model.Formation{
		Name: TestFormationName,
	}
}
