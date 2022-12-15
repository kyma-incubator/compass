package inputvalidationtest

import "strings"

var (
	// String37Long missing godoc
	String37Long = strings.Repeat("a", 37)
	// String101Long missing godoc
	String101Long = strings.Repeat("a", 101)
	// String129Long missing godoc
	String129Long = strings.Repeat("a", 129)
	// String257Long missing godoc
	String257Long = strings.Repeat("a", 257)
	// String2001Long missing godoc
	String2001Long = strings.Repeat("a", 2001)
	// URL257Long missing godoc
	URL257Long = "http://url.com/" + strings.Repeat("a", 242)
)

const (
	// EmptyString missing godoc
	EmptyString = ""
	// ValidName missing godoc
	ValidName = "thi5-1npu7.15-valid"
	// ValidRuntimeNameWithDigit missing godoc
	ValidRuntimeNameWithDigit = "0thi5-1npu7.15_valid"
	// InValidRuntimeNameInvalidCharacters missing godoc
	InValidRuntimeNameInvalidCharacters = "123 456"
	// InvalidName missing godoc
	InvalidName = "0iNvALiD"
	// ValidURL missing godoc
	ValidURL = "https://kyma-project.io"
	// InvalidURL missing godoc
	InvalidURL = "http:/kyma-projectio/path/"
)
