package scope_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/scope/automock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	readScope   = "read"
	writeScope  = "write"
	deleteScope = "delete"
)

func TestHasScope(t *testing.T) {
	t.Run("has all required scopes", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		defer mockRequiredScopesGetter.AssertExpectations(t)
		sut := scope.NewDirective(mockRequiredScopesGetter, &scope.HasScopesErrorProvider{})
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return([]string{readScope, writeScope}, nil).Once()

		next := dummyResolver{}
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})
		// WHEN
		act, err := sut.VerifyScopes(ctx, nil, next.SuccessResolve, fixScopesDefinition())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixNextOutput(), act)
		assert.True(t, next.called)
	})

	t.Run("has insufficient scopes", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		defer mockRequiredScopesGetter.AssertExpectations(t)
		sut := scope.NewDirective(mockRequiredScopesGetter, &scope.HasScopesErrorProvider{})
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return([]string{readScope, writeScope}, nil).Once()
		ctx := scope.SaveToContext(context.TODO(), []string{deleteScope})
		// WHEN
		_, err := sut.VerifyScopes(ctx, nil, nil, fixScopesDefinition())
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required=read;write")
		assert.Contains(t, err.Error(), "actual=delete")
	})

	t.Run("returns error on getting scopes from context", func(t *testing.T) {
		// GIVEN
		sut := scope.NewDirective(nil, &scope.HasScopesErrorProvider{})
		// WHEN
		_, err := sut.VerifyScopes(context.TODO(), nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, "cannot read scopes from context")
	})

	t.Run("returns error on getting required scopes", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		defer mockRequiredScopesGetter.AssertExpectations(t)
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return(nil, fixGivenError()).Once()
		sut := scope.NewDirective(mockRequiredScopesGetter, &scope.HasScopesErrorProvider{})
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})
		// WHEN
		_, err := sut.VerifyScopes(ctx, nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, "while getting required scopes: some error")
	})
}

func TestSanitize(t *testing.T) {
	t.Run("should return sensitive fields when the required scopes are present", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return([]string{readScope, writeScope}, nil).Once()
		defer mockRequiredScopesGetter.AssertExpectations(t)

		next := dummyResolver{}
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})

		dir := scope.NewDirective(mockRequiredScopesGetter, &scope.SanitizeErrorProvider{})
		// WHEN
		act, err := dir.VerifyScopes(ctx, nil, next.SuccessResolve, fixScopesDefinition())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, fixNextOutput(), act)
		assert.True(t, next.called)
	})

	t.Run("should return nil when scopes are insufficient", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return([]string{readScope, writeScope}, nil).Once()
		defer mockRequiredScopesGetter.AssertExpectations(t)

		ctx := scope.SaveToContext(context.TODO(), []string{deleteScope})
		dir := scope.NewDirective(mockRequiredScopesGetter, &scope.SanitizeErrorProvider{})
		// WHEN
		act, err := dir.VerifyScopes(ctx, nil, nil, fixScopesDefinition())
		// THEN
		require.NoError(t, err)
		require.Nil(t, act)
	})

	t.Run("returns error on getting scopes from context", func(t *testing.T) {
		// GIVEN
		dir := scope.NewDirective(nil, &scope.SanitizeErrorProvider{})
		// WHEN
		_, err := dir.VerifyScopes(context.TODO(), nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, apperrors.NoScopesInContextMsg)
	})

	t.Run("returns error on getting required scopes", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return(nil, fixGivenError()).Once()
		defer mockRequiredScopesGetter.AssertExpectations(t)
		dir := scope.NewDirective(mockRequiredScopesGetter, &scope.SanitizeErrorProvider{})
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})
		// WHEN
		_, err := dir.VerifyScopes(ctx, nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, "while getting required scopes: some error")
	})
}

func fixGivenError() error {
	return errors.New("some error")
}

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return fixNextOutput(), nil
}

func fixScopesDefinition() string {
	return "mutations.create.application"
}

func fixNextOutput() string {
	return "nextOutput"
}
