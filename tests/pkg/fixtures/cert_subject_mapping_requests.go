package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixCreateCertificateSubjectMappingRequest(createCertificateSubjectMappingGQLInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: createCertificateSubjectMapping(in: %s) {
    					%s
					}
				}`, createCertificateSubjectMappingGQLInput, testctx.Tc.GQLFieldsProvider.ForCertificateSubjectMapping()))
}

func FixUpdateCertificateSubjectMappingRequest(id, updateCertificateSubjectMappingGQLInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: updateCertificateSubjectMapping(id: %q, in: %s) {
    					%s
					}
				}`, id, updateCertificateSubjectMappingGQLInput, testctx.Tc.GQLFieldsProvider.ForCertificateSubjectMapping()))
}

func FixDeleteCertificateSubjectMappingRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: deleteCertificateSubjectMapping(id: %q) {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForCertificateSubjectMapping()))
}

func FixQueryCertificateSubjectMappingRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: certificateSubjectMapping(id: %q) {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForCertificateSubjectMapping()))
}

func FixQueryCertificateSubjectMappingsRequestWithPageSize(pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: certificateSubjectMappings(first: %d, after:"") {
    					%s
					}
				}`, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForCertificateSubjectMapping())))
}
