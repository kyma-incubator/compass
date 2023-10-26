package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	StatusAPISyncErrorMessage  = "failed to parse request"
	StatusAPIAsyncErrorMessage = "test error"
)

var (
	StatusAPISyncConfigJSON       = str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
	StatusAPIAsyncConfigJSON      = str.Ptr("{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}")
	StatusAPIResetConfigJSON      = str.Ptr("{\"resetKey\":\"resetValue\",\"resetKey2\":{\"resetKey\":\"resetValue2\"}}")
	StatusAPISyncErrorMessageJSON = str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")
	RedirectConfigJSON            = str.Ptr("{\"redirectProperties\":[{\"redirectPropertyName\":\"redirectName\",\"redirectPropertyID\":\"redirectID\"}]}")
	StatusAPISyncError            = &graphql.FormationStatusError{
		Message:   StatusAPISyncErrorMessage,
		ErrorCode: 2,
	}
	StatusAPIAsyncErrorMessageJSON = str.Ptr("{\"error\":{\"message\":\"test error\",\"errorCode\":2}}")
	StatusAPIAsyncError            = &graphql.FormationStatusError{
		Message:   StatusAPIAsyncErrorMessage,
		ErrorCode: 2,
	}
)
