package graphql

import (
	"regexp"

	"github.com/pkg/errors"
)

const (
	shortStringLengthLimit            = 128
	longStringLengthLimit             = 256
	descriptionStringLengthLimit      = 2000
	groupLengthLimit                  = 36
	alpanumericUnderscoreRegexpString = "^[a-zA-Z0-9_]*$"
)

func compileRegexp(r string) (*regexp.Regexp, error) {
	compiledRegexp, err := regexp.Compile(r)
	if err != nil {
		return nil, errors.Wrapf(err, "while compiling regexp [%s]", r)
	}
	return compiledRegexp, nil
}
