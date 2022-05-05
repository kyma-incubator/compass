/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func ListAutomaticScenarioAssignmentsWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string) graphql.AutomaticScenarioAssignmentPage {
	assignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	req := FixAutomaticScenarioAssignmentsRequest()
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, req, &assignmentsPage)
	require.NoError(t, err)
	return assignmentsPage
}
