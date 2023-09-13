package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	StatusAPISyncErrorMessage = "failed to parse request"
)

var (
	StatusAPISyncConfigJSON       = str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
	StatusAPIAsyncConfigJSON      = str.Ptr("{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}")
	StatusAPISyncErrorMessageJSON = str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")
	StatusAPISyncError            = &graphql.FormationStatusError{
		Message:   "failed to parse request",
		ErrorCode: 2,
	}
	StatusAPIAsyncErrorMessageJSON = str.Ptr("{\"error\":{\"message\":\"test error\",\"errorCode\":2}}")
)
