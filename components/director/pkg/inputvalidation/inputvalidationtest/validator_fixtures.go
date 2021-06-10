package inputvalidationtest

import "strings"

var (
	String37Long   = strings.Repeat("a", 37)
	String129Long  = strings.Repeat("a", 129)
	String257Long  = strings.Repeat("a", 257)
	String2001Long = strings.Repeat("a", 2001)
	URL257Long     = "http://url.com/" + strings.Repeat("a", 242)
)

const (
	EmptyString                         = ""
	ValidName                           = "thi5-1npu7.15-valid"
	ValidRuntimeNameWithDigit           = "0thi5-1npu7.15_valid"
	InValidRuntimeNameTooLong           = "123aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaааа"
	InValidRuntimeNameInvalidCharacters = "123 456"
	InvalidName                         = "0iNvALiD"
	ValidURL                            = "https://kyma-project.io"
	InvalidURL                          = "http:/kyma-projectio/path/"
)
