package release

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-project/control-plane/components/provisioner/internal/installation/release/mocks"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/persistence/dberrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	onDemandVersion = "master-1bcdef"
)

func TestOnDemand_GetReleaseByVersion(t *testing.T) {

	release := model.Release{
		Id:            "abcd-efgh",
		Version:       onDemandVersion,
		TillerYAML:    "tiller",
		InstallerYAML: "installer",
	}

	t.Run("should get on demand release from database", func(t *testing.T) {
		// given
		repo := &mocks.Repository{}
		repo.On("GetReleaseByVersion", onDemandVersion).Return(release, nil)

		onDemand := NewOnDemandWrapper(nil, repo)

		// when
		onDemandRelease, err := onDemand.GetReleaseByVersion(onDemandVersion)
		require.NoError(t, err)

		// then
		assert.Equal(t, release, onDemandRelease)
	})

	t.Run("should download and save on demand release", func(t *testing.T) {
		for _, testCase := range []struct {
			description string
			httpClient  *http.Client
			release     model.Release
		}{
			{
				description: "with Tiller",
				httpClient: newTestClient(func(req *http.Request) *http.Response {
					if strings.HasSuffix(req.URL.String(), "kyma-installer-cluster.yaml") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString("installer")),
						}
					}
					if strings.HasSuffix(req.URL.String(), "tiller.yaml") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString("tiller")),
						}
					}
					return &http.Response{
						StatusCode: http.StatusBadRequest,
					}
				}),
				release: model.Release{
					Version:       onDemandVersion,
					TillerYAML:    "tiller",
					InstallerYAML: "installer",
				},
			},
			{
				description: "without Tiller",
				httpClient: newTestClient(func(req *http.Request) *http.Response {
					if strings.HasSuffix(req.URL.String(), "kyma-installer-cluster.yaml") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString("installer")),
						}
					}
					if strings.HasSuffix(req.URL.String(), "tiller.yaml") {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       ioutil.NopCloser(bytes.NewBufferString("404 not found")),
						}
					}
					return &http.Response{
						StatusCode: http.StatusBadRequest,
					}
				}),
				release: model.Release{
					Version:       onDemandVersion,
					InstallerYAML: "installer",
				},
			},
		} {
			t.Run(testCase.description, func(t *testing.T) {
				// given
				fileDownloader := NewFileDownloader(testCase.httpClient)

				repo := &mocks.Repository{}
				repo.On("GetReleaseByVersion", onDemandVersion).Return(model.Release{}, dberrors.NotFound("error"))
				repo.On("SaveRelease", testCase.release).Run(func(args mock.Arguments) {
					rel, ok := args.Get(0).(model.Release)
					require.True(t, ok)
					(&rel).Id = "abcd-efgh"
				}).Return(release, nil)

				onDemand := NewOnDemandWrapper(fileDownloader, repo)

				// when
				onDemandRelease, err := onDemand.GetReleaseByVersion(onDemandVersion)
				require.NoError(t, err)

				// then
				assert.Equal(t, release, onDemandRelease)
			})
		}
	})
}

func TestOnDemand_GetReleaseByVersion_Error(t *testing.T) {

	t.Run("should return error when failed to get release from database", func(t *testing.T) {
		// given
		repo := &mocks.Repository{}
		repo.On("GetReleaseByVersion", onDemandVersion).Return(model.Release{}, dberrors.Internal("error"))

		onDemand := NewOnDemandWrapper(nil, repo)

		// when
		_, err := onDemand.GetReleaseByVersion(onDemandVersion)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when non on-demand version not found in database", func(t *testing.T) {
		// given
		repo := &mocks.Repository{}
		repo.On("GetReleaseByVersion", "not-supported-version").Return(model.Release{}, dberrors.NotFound("error"))

		onDemand := NewOnDemandWrapper(nil, repo)

		// when
		_, err := onDemand.GetReleaseByVersion("not-supported-version")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to save release", func(t *testing.T) {
		// given
		httpClient := newTestClient(func(req *http.Request) *http.Response {
			if strings.HasSuffix(req.URL.String(), "kyma-installer-cluster.yaml") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("installer")),
				}
			}
			if strings.HasSuffix(req.URL.String(), "tiller.yaml") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("tiller")),
				}
			}
			return &http.Response{
				StatusCode: http.StatusBadRequest,
			}
		})
		fileDownloader := NewFileDownloader(httpClient)

		repo := &mocks.Repository{}
		repo.On("GetReleaseByVersion", onDemandVersion).Return(model.Release{}, dberrors.NotFound("error"))
		repo.On("SaveRelease", mock.AnythingOfType("model.Release")).Return(model.Release{}, dberrors.Internal("error"))

		onDemand := NewOnDemandWrapper(fileDownloader, repo)

		// when
		_, err := onDemand.GetReleaseByVersion(onDemandVersion)

		// then
		require.Error(t, err)
	})

	for _, testCase := range []struct {
		description string
		httpClient  *http.Client
	}{
		{
			description: "should return error when failed to download tiller",
			httpClient: newTestClient(func(req *http.Request) *http.Response {
				if strings.HasSuffix(req.URL.String(), "kyma-installer-cluster.yaml") {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString("installer")),
					}
				}
				if strings.HasSuffix(req.URL.String(), "tiller.yaml") {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}
				}
				return &http.Response{
					StatusCode: http.StatusBadRequest,
				}
			}),
		},
		{
			description: "should return error when failed to download installer",
			httpClient: newTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				}
			}),
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			repo := &mocks.Repository{}
			repo.On("GetReleaseByVersion", onDemandVersion).Return(model.Release{}, dberrors.NotFound("error"))
			fileDownloader := NewFileDownloader(testCase.httpClient)

			onDemand := NewOnDemandWrapper(fileDownloader, repo)

			// when
			_, err := onDemand.GetReleaseByVersion(onDemandVersion)

			// then
			require.Error(t, err)
		})
	}

}
