package artifacts

import (
	"context"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

type Repository interface {
	GetArtifacts(version string) (ReleaseArtifacts, dberrors.Error)
	SaveArtifacts(artifacts ReleaseArtifacts) (ReleaseArtifacts, dberrors.Error)
}

func NewArtifactsDownloader(repository Repository, latestReleases int, includePreReleases bool) *artifactsDownloader {
	return &artifactsDownloader{
		repository: repository,
	}
}

type artifactsDownloader struct {
	repository         Repository
	latestReleases     int
	includePreReleases bool
}

func (ad artifactsDownloader) FetchPeriodically(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := ad.fetchLatestReleases()
			if err != nil {
				// TODO - log some error
			}

			//time.Sleep() // TODO - sleep if no error
		}
	}

}

func (ad artifactsDownloader) fetchLatestReleases() error {

	// TODO - fetch n latest releases
	// TODO - save releases to DB

	return nil
}
