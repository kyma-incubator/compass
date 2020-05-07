package model_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntimeEventingConfiguration(t *testing.T) {
	testCases := []struct {
		Name          string
		RawURL        string
		ExpectedError error
	}{
		{
			Name:   "Valid RuntimeEventing",
			RawURL: "https://eventing.runtime",
		},
		{
			Name:   "Valid RuntimeEventing - empty rawURL",
			RawURL: "",
		},
		{
			Name:          "Invalid rawURL",
			RawURL:        "::",
			ExpectedError: errors.New("parse \"::\": missing protocol scheme"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			//WHEN
			config, err := model.NewRuntimeEventingConfiguration(testCase.RawURL)

			//THEN
			if testCase.ExpectedError == nil {
				require.NotNil(t, config)
				require.Equal(t, testCase.RawURL, config.DefaultURL.String())
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestNewApplicationEventingConfiguration(t *testing.T) {
	validURL, err := url.Parse("https://eventing.runtime")
	require.NoError(t, err)
	require.NotNil(t, validURL)

	t.Run("Valid", func(t *testing.T) {
		//GIVEN
		expectedEventURL := "https://eventing.runtime/app.name-super/v1/events"

		//WHEN
		config, err := model.NewApplicationEventingConfiguration(*validURL, "app.name-super")

		//THEN
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Equal(t, expectedEventURL, config.DefaultURL.String())

	})

	t.Run("Valid - empty runtimeEventURL", func(t *testing.T) {
		//GIVEN
		emptyURL, err := url.Parse("")
		require.NoError(t, err)
		require.NotNil(t, emptyURL)

		//WHEN
		config, err := model.NewApplicationEventingConfiguration(*emptyURL, "")

		//THEN
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "", config.DefaultURL.String())
	})
}

func TestNewEmptyApplicationEventingConfig(t *testing.T) {
	//WHEN
	config, err := model.NewEmptyApplicationEventingConfig()
	//THEN
	require.NoError(t, err)
	assert.Equal(t, "", config.DefaultURL.String())
}
