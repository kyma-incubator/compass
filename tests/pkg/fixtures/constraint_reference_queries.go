package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AttachConstraintToFormationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, constraintID, constraintName, formationTemplateID, formationTemplateName string) *graphql.ConstraintReference {
	t.Logf("Attaching formation constraint with ID: %s and name: %s to template with ID: %s and name: %s", constraintID, constraintName, formationTemplateID, formationTemplateName)
	createRequest := FixAttachConstraintToFormationTemplateRequest(constraintID, formationTemplateID)
	constraintReference := graphql.ConstraintReference{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &constraintReference))

	return &constraintReference
}

func DetachConstraintFromFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, constraintID, formationTemplateID string) *graphql.ConstraintReference {
	detachRequest := FixDetachConstraintFromFormationTemplateRequest(constraintID, formationTemplateID)
	constraintReference := graphql.ConstraintReference{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, detachRequest, &constraintReference)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &constraintReference
}

func DetachConstraintFromFormationTemplateNoCheckError(ctx context.Context, gqlClient *gcli.Client, constraintID, formationTemplateID string) *graphql.ConstraintReference {
	detachRequest := FixDetachConstraintFromFormationTemplateRequest(constraintID, formationTemplateID)
	constraintReference := graphql.ConstraintReference{}
	_ = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, detachRequest, &constraintReference)
	return &constraintReference
}
