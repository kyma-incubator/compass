package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
)

const (
	//TODO Think if some of those values should be moved to chart values
	LongInterval  = 1 * time.Hour
	ShortInterval = 1 * time.Minute

	Timeout = 5 * time.Second

	releaseFetchURL   = "https://api.github.com/repos/kyma-project/kyma/releases"
	installerYAMLName = "kyma-installer-cluster.yaml"
	tillerFormat      = "https://raw.githubusercontent.com/kyma-project/kyma/release-%s/installation/resources/tiller.yaml"
)

func NewArtifactsDownloader(repository Repository, latestReleases int, includePreReleases bool, client *http.Client, log *logrus.Entry) *artifactsDownloader {
	return &artifactsDownloader{
		repository:         repository,
		latestReleases:     latestReleases,
		includePreReleases: includePreReleases,
		httpClient:         client,
		log:                log,
	}
}

type artifactsDownloader struct {
	repository         Repository
	latestReleases     int
	includePreReleases bool
	httpClient         *http.Client
	log                *logrus.Entry
}

func (ad artifactsDownloader) FetchPeriodically(ctx context.Context, shortInterval, longInterval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := ad.fetchLatestReleases()
			if err != nil {
				ad.log.Errorf("Error during release fetch: %s", err.Error())
				time.Sleep(shortInterval)
			} else {
				time.Sleep(longInterval)
			}
		}
	}
}

func (ad artifactsDownloader) fetchLatestReleases() error {
	releases, err := ad.fetchReleases()

	if err != nil {
		return err
	}

	if !ad.includePreReleases {
		releases = filterPreReleases(releases)
	}

	releases = getLatestReleases(releases, ad.latestReleases)

	return ad.save(releases)
}

func (ad artifactsDownloader) fetchReleases() ([]model.GithubRelease, error) {
	body, err := ad.sendRequest(releaseFetchURL)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	var releases []model.GithubRelease

	err = json.NewDecoder(body).Decode(&releases)

	if err != nil {
		return nil, err
	}

	return releases, nil
}

func (ad artifactsDownloader) save(releases []model.GithubRelease) error {
	for _, release := range releases {
		artifacts, err := ad.buildRelease(release)
		if err != nil {
			return err
		}

		exists, err := ad.repository.ReleaseExists(artifacts.Version)

		if err != nil {
			return err
		}

		if !exists {
			_, err = ad.repository.SaveRelease(artifacts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ad artifactsDownloader) buildRelease(release model.GithubRelease) (model.Release, error) {
	var installerURL string
	for _, a := range release.Assets {
		if a.Name == installerYAMLName {
			installerURL = a.Url
			break
		}
	}

	if "" == installerURL {
		return model.Release{}, errors.New(fmt.Sprintf("Release %d does not contain installer yaml", release.Id))
	}

	tillerURL := buildTillerURL(release.Name)

	installerYAML, err := ad.downloadYAML(installerURL)
	tillerYAML, err := ad.downloadYAML(tillerURL)

	if err != nil {
		return model.Release{}, err
	}

	return model.Release{
		Version:       release.Name,
		TillerYAML:    tillerYAML,
		InstallerYAML: installerYAML,
	}, nil
}

func (ad artifactsDownloader) downloadYAML(url string) (string, error) {
	body, err := ad.sendRequest(url)

	if err != nil {
		return "", err
	}

	defer body.Close()

	bytes, err := ioutil.ReadAll(body)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (ad artifactsDownloader) sendRequest(url string) (io.ReadCloser, error) {
	resp, err := ad.httpClient.Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Received unexpected http status %d", resp.StatusCode))
	}

	return resp.Body, nil
}

func filterPreReleases(releases []model.GithubRelease) []model.GithubRelease {
	var filtered []model.GithubRelease

	for _, r := range releases {
		if !r.Prerelease {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func getLatestReleases(releases []model.GithubRelease, latestReleases int) []model.GithubRelease {
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Id > releases[j].Id
	})

	if len(releases) < latestReleases {
		return releases
	}

	return releases[0:latestReleases]
}

func buildTillerURL(releaseName string) string {
	return fmt.Sprintf(tillerFormat, releaseName[0:3])
}
