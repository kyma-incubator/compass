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
	Match(params ResourceArguments) (string, error)
	MatchMultiple(params ResourceArguments) []string
	GetType() string
}

// ResourceArguments holds all specific arguments needed for matching resources and provides the specific URL Path of each one
//
//go:generate mockery --name=ResourceArguments --output=automock --outpkg=automock --case=underscore --disable-version-string
type ResourceArguments interface {
	GetURLPath() string
}

// ResourceRequestBody represents a request body for each SM resource needed on creation
//
//go:generate mockery --name=ResourceRequestBody --output=automock --outpkg=automock --case=underscore --disable-version-string
type ResourceRequestBody interface {
	GetResourceName() string
}
