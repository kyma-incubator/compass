package inputvalidationtest

import "strings"

var (
	String37Long = strings.Repeat("a", 37)
	URL257Long   = "http://url.com/" + strings.Repeat("a", 242)
)

const (
	EmptyString = ""
	ValidName   = "thi5-1npu7.15-valid"
	ValidURL    = "https://kyma-project.io"
	InvalidURL  = "http:/kyma-projectio/path/"
)
