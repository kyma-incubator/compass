package release

import (
	"bytes"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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

		repo := &mocks.Repository{}
		repo.On("GetReleaseByVersion", onDemandVersion).Return(model.Release{}, dberrors.NotFound("error"))
		repo.On("SaveRelease", mock.AnythingOfType("model.Release")).Run(func(args mock.Arguments) {
			rel, ok := args.Get(0).(model.Release)
			require.True(t, ok)
			(&rel).Id = "abcd-efgh"
		}).Return(release, nil)

		onDemand := NewOnDemandWrapper(httpClient, repo)

		// when
		onDemandRelease, err := onDemand.GetReleaseByVersion(onDemandVersion)
		require.NoError(t, err)

		// then
		assert.Equal(t, release, onDemandRelease)
	})

}
