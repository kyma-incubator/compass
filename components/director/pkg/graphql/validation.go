package graphql

import (
	"regexp"
)

const (
	shortStringLengthLimit             = 128
	longStringLengthLimit              = 256
	longLongStringLengthLimit          = 512
	descriptionStringLengthLimit       = 2000
	jsonPathStringLengthLimit          = 2000
	appNameLengthLimit                 = 100
	groupLengthLimit                   = 100
	alphanumericUnderscoreRegexpString = "^[a-zA-Z0-9_]*$"
)

var alphanumericUnderscoreRegexp = regexp.MustCompile(alphanumericUnderscoreRegexpString)
