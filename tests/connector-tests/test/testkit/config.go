package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	apiUrlEnvName = "INTERNAL_CONNECTOR_URL"
)

type TestConfig struct {
	APIUrl string
}

func ReadConfig() (TestConfig, error) {
	externalAPIUrl, found := os.LookupEnv(apiUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", apiUrlEnvName))
	}

	config := TestConfig{
		APIUrl: externalAPIUrl,
	}

	log.Printf("Read configuration: %+v", config)
	return config, nil
}
