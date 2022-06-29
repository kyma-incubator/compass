package testingx

import (
	"regexp"
	"testing"
)

const skipRegexEnvVar = "SKIP_TESTS_REGEX"

type T struct {
	*testing.T
}

func NewT(t *testing.T) *T {
	return &T{
		T: t,
	}
}

func (t *T) Run(name string, f func(t *testing.T)) bool {
	newF := f

	pattern := "TestRefetchAPISpecDifferentSpec"
	if len(pattern) > 0 {
		newF = func(t *testing.T) {
			match, err := regexp.MatchString(pattern, t.Name())
			if err != nil {
				t.Fatalf("An error occured while parsing skip regex: %s", pattern)
			}
			if !match {
				t.Skipf("Skipping test... Reason: test name %s doesn't match pattern %s", t.Name(), pattern)
			}
			f(t)
		}
	}

	return t.T.Run(name, newF)
}
