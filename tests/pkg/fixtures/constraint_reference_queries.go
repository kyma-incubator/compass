package fixtures

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AttachConstraintToFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, constraintID, formationTemplateID string) *graphql.ConstraintReference {
	createRequest := FixAttachConstraintToFormationTemplateRequest(constraintID, formationTemplateID)
	constraintReference := graphql.ConstraintReference{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &constraintReference))

	return &constraintReference
}

func DetachConstraintFromFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, constraintID, formationTemplateID string) *graphql.ConstraintReference {
	deleteRequest := FixDetachConstraintFromFormationTemplateRequest(constraintID, formationTemplateID)
	constraintReference := graphql.ConstraintReference{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &constraintReference)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &constraintReference
}
