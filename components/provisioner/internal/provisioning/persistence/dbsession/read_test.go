package dbsession

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseToKymaConfig(t *testing.T) {

	kymaConfigId := "abc-kyma-def"
	runtimeId := "abc-runtime-def"
	releaseId := "abc-release-def"
	version := "18.0.0"
	tillerYaml := "tiller"
	installerYaml := "installer"

	newComponentDTO := func(id, name, namespace string, order int) kymaComponentConfigDTO {
		return kymaComponentConfigDTO{
			ID:                  id,
			KymaConfigID:        kymaConfigId,
			ReleaseID:           releaseId,
			Version:             version,
			TillerYAML:          tillerYaml,
			InstallerYAML:       installerYaml,
			Component:           name,
			Namespace:           namespace,
			ComponentOrder:      order,
			ClusterID:           runtimeId,
			GlobalConfiguration: []byte("{}"),
			Configuration:       []byte("{}"),
		}
	}

	for _, testCase := range []struct {
		description    string
		kymaConfigDTO  kymaConfigDTO
		expectedConfig model.KymaConfig
	}{
		{
			description: "should parse using component order",
			kymaConfigDTO: kymaConfigDTO{
				newComponentDTO("comp-3", "even-less-essential", "core", 3),
				newComponentDTO("comp-1", "essential", "core", 1),
				newComponentDTO("comp-2", "less-essential", "other", 2),
			},
			expectedConfig: model.KymaConfig{
				ID: kymaConfigId,
				Release: model.Release{
					Id:            releaseId,
					Version:       version,
					TillerYAML:    tillerYaml,
					InstallerYAML: installerYaml,
				},
				Components: []model.KymaComponentConfig{
					{
						ID:             "comp-1",
						Component:      "essential",
						Namespace:      "core",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 1,
						KymaConfigID:   kymaConfigId,
					},
					{
						ID:             "comp-2",
						Component:      "less-essential",
						Namespace:      "other",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 2,
						KymaConfigID:   kymaConfigId,
					},
					{
						ID:             "comp-3",
						Component:      "even-less-essential",
						Namespace:      "core",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 3,
						KymaConfigID:   kymaConfigId,
					},
				},
				GlobalConfiguration: model.Configuration{},
				ClusterID:           runtimeId,
			},
		},
		{
			description: "should parse in order of reed if component order is equal",
			kymaConfigDTO: kymaConfigDTO{
				newComponentDTO("comp-3", "even-less-essential", "core", 0),
				newComponentDTO("comp-1", "essential", "core", 0),
				newComponentDTO("comp-2", "less-essential", "other", 0),
			},
			expectedConfig: model.KymaConfig{
				ID: kymaConfigId,
				Release: model.Release{
					Id:            releaseId,
					Version:       version,
					TillerYAML:    tillerYaml,
					InstallerYAML: installerYaml,
				},
				Components: []model.KymaComponentConfig{
					{
						ID:             "comp-3",
						Component:      "even-less-essential",
						Namespace:      "core",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 0,
						KymaConfigID:   kymaConfigId,
					},
					{
						ID:             "comp-1",
						Component:      "essential",
						Namespace:      "core",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 0,
						KymaConfigID:   kymaConfigId,
					},
					{
						ID:             "comp-2",
						Component:      "less-essential",
						Namespace:      "other",
						SourceURL:      "",
						Configuration:  model.Configuration{},
						ComponentOrder: 0,
						KymaConfigID:   kymaConfigId,
					},
				},
				GlobalConfiguration: model.Configuration{},
				ClusterID:           runtimeId,
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			kymaConfig, err := testCase.kymaConfigDTO.parseToKymaConfig(runtimeId)
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedConfig, kymaConfig)
		})
	}

}
