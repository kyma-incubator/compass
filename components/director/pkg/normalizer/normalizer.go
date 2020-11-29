package normalizer

import (
	"regexp"
	"strings"
)

// defaultNormalizationPrefix is a fixed string used as prefix during default normalization
const defaultNormalizationPrefix = "mp-"

// Func represents the interface of a normalization function used to sanitize a given name string
type Func func(string) string

// New constructs the appropriate normalization function based on whether normalization is enabled or not
func New(enableNormalization bool) Func {
	if enableNormalization {
		return DefaultNormalizer
	}

	return NoopNormalizer
}

// NoopNormalizer is a noop normalization function which doesn't apply any normalization rules to a given input string
func NoopNormalizer(name string) string {
	return name
}

// DefaultNormalizer is the default normalization function used when normalization function
func DefaultNormalizer(name string) string {
	var normalizedName = defaultNormalizationPrefix + name

	normalizedName = strings.ToLower(normalizedName)
	normalizedName = regexp.MustCompile("[^-a-z0-9]").ReplaceAllString(normalizedName, "-")
	normalizedName = regexp.MustCompile("-{2,}").ReplaceAllString(normalizedName, "-")
	normalizedName = regexp.MustCompile("-$").ReplaceAllString(normalizedName, "")

	return normalizedName
}
