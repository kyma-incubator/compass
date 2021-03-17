package sanitize_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/sanitize"
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

func TestSanitize(t *testing.T) {
	t.Run("should return sensitive fields when the required scopes are present", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return([]string{readScope, writeScope}, nil).Once()
		defer mockRequiredScopesGetter.AssertExpectations(t)

		next := dummyResolver{}
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})

		dir := sanitize.NewDirective(mockRequiredScopesGetter)
		// WHEN
		act, err := dir.Sanitize(ctx, nil, next.SuccessResolve, fixScopesDefinition())
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
		dir := sanitize.NewDirective(mockRequiredScopesGetter)
		// WHEN
		act, err := dir.Sanitize(ctx, nil, nil, fixScopesDefinition())
		// THEN
		require.NoError(t, err)
		require.Nil(t, act)
	})

	t.Run("returns error on getting scopes from context", func(t *testing.T) {
		// GIVEN
		dir := sanitize.NewDirective(nil)
		// WHEN
		_, err := dir.Sanitize(context.TODO(), nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, apperrors.NoScopesInContextMsg)
	})

	t.Run("returns error on getting required scopes", func(t *testing.T) {
		// GIVEN
		mockRequiredScopesGetter := &automock.ScopesGetter{}
		mockRequiredScopesGetter.On("GetRequiredScopes", fixScopesDefinition()).Return(nil, fixGivenError()).Once()
		defer mockRequiredScopesGetter.AssertExpectations(t)
		dir := sanitize.NewDirective(mockRequiredScopesGetter)
		ctx := scope.SaveToContext(context.TODO(), []string{readScope, writeScope, deleteScope})
		// WHEN
		_, err := dir.Sanitize(ctx, nil, nil, fixScopesDefinition())
		// THEN
		assert.EqualError(t, err, "while getting required scopes: some error")
	})
}

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return fixNextOutput(), nil
}

func fixGivenError() error {
	return errors.New("some error")
}

func fixScopesDefinition() string {
	return "mutations.create.application"
}

func fixNextOutput() string {
	return "nextOutput"
}
