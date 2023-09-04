package resources

// Resource provides common capabilities for each SM resources that is used by the SM client
//
//go:generate mockery --name=Resource --output=automock --outpkg=automock --case=underscore --disable-version-string
type Resource interface {
	GetResourceID() string
	GetResourceType() string
	GetResourceURLPath() string
}

// Resources provides common capabilities for all the SM resources that are used by the SM client
//
//go:generate mockery --name=Resources --output=automock --outpkg=automock --case=underscore --disable-version-string
type Resources interface {
	Match(params ResourceMatchParameters) (string, error)
	MatchMultiple(params ResourceMatchParameters) ([]string, error)
	GetType() string
}

// ResourceMatchParameters holds all specific parameters needed for matching resources and provides the specific URL Path of each one
//
//go:generate mockery --name=ResourceMatchParameters --output=automock --outpkg=automock --case=underscore --disable-version-string
type ResourceMatchParameters interface {
	GetURLPath() string
}

// ResourceRequestBody represents a request body for each SM resource needed on creation
//
//go:generate mockery --name=ResourceRequestBody --output=automock --outpkg=automock --case=underscore --disable-version-string
type ResourceRequestBody interface {
	GetResourceName() string
}
