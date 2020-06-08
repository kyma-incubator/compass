package release

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testRelease struct {
	githubRelease   model.GithubRelease
	expectedRelease model.Release
}

func TestArtifactsDownloader_FetchPeriodically(t *testing.T) {

	longInterval := 3 * time.Second
	shortInterval := 1 * time.Second

	installerURL := "https://github.com/kyma-project/kyma/testReleases/download/version/kyma-installer-cluster.yaml"
	installerContent := "some installer content"
	tillerContent := "some tiller content"

	//given
	testReleases := []testRelease{
		{
			githubRelease: model.GithubRelease{
				Id:         100,
				Name:       "1.7",
				Prerelease: false,
				Assets: []model.Asset{
					{Name: "kyma-installer-cluster.yaml", Url: installerURL},
					{Name: "some-other-asset", Url: ""},
				}},
			expectedRelease: model.Release{
				Version:       "1.7",
				TillerYAML:    tillerContent,
				InstallerYAML: installerContent,
			},
		},
		{
			githubRelease: model.GithubRelease{
				Id:         101,
				Name:       "1.8",
				Prerelease: false,
				Assets: []model.Asset{
					{Name: "kyma-installer-cluster.yaml", Url: installerURL},
					{Name: "ya-ya-cthulhu-fhtagn", Url: ""},
				}},
			expectedRelease: model.Release{
				Version:       "1.8",
				TillerYAML:    tillerContent,
				InstallerYAML: installerContent,
			},
		},
		{
			githubRelease: model.GithubRelease{
				Id:         102,
				Name:       "1.9-rc2",
				Prerelease: true,
				Assets: []model.Asset{
					{Name: "kyma-installer-cluster.yaml", Url: installerURL},
					{Name: "ya-ya-cthulhu-fhtagn", Url: ""},
				}},
			expectedRelease: model.Release{
				Version:       "1.9-rc2",
				TillerYAML:    "",
				InstallerYAML: installerContent,
			},
		},
	}

	t.Run("Should fetch testReleases", func(t *testing.T) {
		//given
		releases := []model.GithubRelease{testReleases[0].githubRelease, testReleases[1].githubRelease}

		client := newMockClient(t, releases, installerURL, installerContent, tillerContent)
		fileDownloader := NewFileDownloader(client)

		repository := &mocks.Repository{}
		repository.On("ReleaseExists", mock.Anything).Return(false, nil)

		repository.On("SaveRelease", testReleases[0].expectedRelease).Return(testReleases[0].expectedRelease, nil)
		repository.On("SaveRelease", testReleases[1].expectedRelease).Return(testReleases[1].expectedRelease, nil)

		entry := logrus.WithField("Component", "ArtifactsDownloaderTests")

		downloader := NewArtifactsDownloader(repository, 3, true, client, fileDownloader, entry)

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		//when
		downloader.FetchPeriodically(ctx, shortInterval, longInterval)

		//then
		repository.AssertExpectations(t)
	})

	t.Run("Should fetch testReleases without prereleases", func(t *testing.T) {
		//given
		releases := []model.GithubRelease{testReleases[0].githubRelease, testReleases[1].githubRelease, testReleases[2].githubRelease}

		client := newMockClient(t, releases, installerURL, installerContent, tillerContent)
		fileDownloader := NewFileDownloader(client)

		repository := &mocks.Repository{}
		repository.On("ReleaseExists", mock.Anything).Return(false, nil)

		repository.On("SaveRelease", testReleases[0].expectedRelease).Return(testReleases[0].expectedRelease, nil)
		repository.On("SaveRelease", testReleases[1].expectedRelease).Return(testReleases[1].expectedRelease, nil)

		entry := logrus.WithField("Component", "ArtifactsDownloaderTests")

		downloader := NewArtifactsDownloader(repository, 3, false, client, fileDownloader, entry)

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		//when
		downloader.FetchPeriodically(ctx, shortInterval, longInterval)

		//then
		repository.AssertExpectations(t)
	})

	t.Run("Should fetch only one, latest, release", func(t *testing.T) {
		releases := []model.GithubRelease{testReleases[0].githubRelease, testReleases[1].githubRelease, testReleases[2].githubRelease}

		client := newMockClient(t, releases, installerURL, installerContent, tillerContent)
		fileDownloader := NewFileDownloader(client)

		expectedReleaseThree := model.Release{
			Version:       "1.9-rc2",
			TillerYAML:    tillerContent,
			InstallerYAML: installerContent,
		}

		repository := &mocks.Repository{}
		repository.On("ReleaseExists", mock.Anything).Return(false, nil)

		repository.On("SaveRelease", expectedReleaseThree).Return(expectedReleaseThree, nil)

		entry := logrus.WithField("Component", "ArtifactsDownloaderTests")

		downloader := NewArtifactsDownloader(repository, 1, true, client, fileDownloader, entry)

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		//when
		downloader.FetchPeriodically(ctx, shortInterval, longInterval)

		//then
		repository.AssertExpectations(t)
	})

	t.Run("Should fetch release without Tiller", func(t *testing.T) {
		//given
		releases := []model.GithubRelease{testReleases[2].githubRelease}

		client := newMockClient(t, releases, installerURL, installerContent, "")
		fileDownloader := NewFileDownloader(client)

		repository := &mocks.Repository{}
		repository.On("ReleaseExists", mock.Anything).Return(false, nil)

		repository.On("SaveRelease", testReleases[2].expectedRelease).Return(testReleases[2].expectedRelease, nil)

		entry := logrus.WithField("Component", "ArtifactsDownloaderTests")

		downloader := NewArtifactsDownloader(repository, 3, true, client, fileDownloader, entry)

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		//when
		downloader.FetchPeriodically(ctx, shortInterval, longInterval)

		//then
		repository.AssertExpectations(t)
	})

	t.Run("Should not save release if already exists in database", func(t *testing.T) {
		releases := []model.GithubRelease{testReleases[2].githubRelease}

		client := newMockClient(t, releases, installerURL, installerContent, tillerContent)
		fileDownloader := NewFileDownloader(client)

		repository := &mocks.Repository{}
		repository.On("ReleaseExists", "1.9-rc2").Return(true, nil)

		entry := logrus.WithField("Component", "ArtifactsDownloaderTests")

		downloader := NewArtifactsDownloader(repository, 1, true, client, fileDownloader, entry)

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, 5*time.Second)

		//when
		downloader.FetchPeriodically(ctx, shortInterval, longInterval)

		//then
		repository.AssertExpectations(t)
	})
}

func newMockClient(t *testing.T, releases []model.GithubRelease, installerURL, installerContent, tillerContent string) *http.Client {
	return newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == releaseFetchURL {

			content, err := json.Marshal(releases)

			require.NoError(t, err)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(content)),
			}
		}

		if req.URL.String() == installerURL {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(installerContent)),
			}
		}

		if strings.Contains(req.URL.String(), "tiller.yaml") {
			if tillerContent == "" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewBufferString("404 not found")),
				}
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(tillerContent)),
			}
		}

		return &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	})
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(rtFunc RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(rtFunc),
	}
}
