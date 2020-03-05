package release

import (
	"sync"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

// Aggregator provides functionality for aggregating release providers and operating on them
type Aggregator struct {
	releaseProviders []ReadRepository
	mu               sync.RWMutex
}

// NewAggregator returns new instance of Aggregator
func NewAggregator(initReleaseProviders ...ReadRepository) *Aggregator {
	return &Aggregator{
		releaseProviders: initReleaseProviders,
	}
}

// GetReleaseByVersion checks if the given version is recognized by previously registered
// release providers. If yes then returns release details
func (a *Aggregator) GetReleaseByVersion(version string) (model.Release, dberrors.Error) {
	for i := range a.releaseProviders {
		found, err := a.releaseProviders[i].ReleaseExists(version)
		if err != nil {
			return model.Release{}, dberrors.Internal("Failed to select release provider for version %s: %s", version, err)
		}
		if found {
			return a.releaseProviders[i].GetReleaseByVersion(version)
		}
	}

	return model.Release{}, dberrors.Internal("Failed to find release provider for version %s", version)
}

// ReleaseExists executes registered providers as long as some of the provider recognizes the given release version
func (a *Aggregator) ReleaseExists(version string) (bool, dberrors.Error) {
	for i := range a.releaseProviders {
		found, err := a.releaseProviders[i].ReleaseExists(version)
		if err != nil {
			return false, err
		}
		if found {
			return found, nil
		}
	}

	return false, nil
}

// Register adds new release provider
func (a *Aggregator) Register(releaseProvider ReadRepository) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releaseProviders = append(a.releaseProviders, releaseProvider)
}
