package graphql

import (
	"regexp"
)

const (
	shortStringLengthLimit             = 128
	longStringLengthLimit              = 256
	descriptionStringLengthLimit       = 2000
	groupLengthLimit                   = 36
	alphanumericUnderscoreRegexpString = "^[a-zA-Z0-9_]*$"
)

var alphanumericUnderscoreRegexp = regexp.MustCompile(alphanumericUnderscoreRegexpString)
