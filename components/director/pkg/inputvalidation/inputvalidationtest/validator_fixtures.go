package inputvalidationtest

import "strings"

var (
	String37Long  = strings.Repeat("a", 37)
	String129Long = strings.Repeat("a", 129)
	String257Long = strings.Repeat("a", 257)
	URL257Long    = "http://url.com/" + strings.Repeat("a", 242)
)

const (
	EmptyString = ""
	ValidName   = "thi5-1npu7.15-valid"
	InvalidName = "0iNvALiD"
	ValidUUID   = "d221a3d4-258d-430d-ad5e-5729aa400ffc"
	InvalidUUID = "zxc4-258d430dad5e-5fc"
	ValidURL    = "https://kyma-project.io"
	InvalidURL  = "http:/kyma-projectio/path/"
)
