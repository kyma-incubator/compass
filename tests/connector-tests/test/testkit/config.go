package testkit

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	internalAPIUrlEnvName = "INTERNAL_API_URL"
	externalAPIUrlEnvName = "EXTERNAL_API_URL"
)

type TestConfig struct {
	InternalAPIUrl string
	ExternalAPIUrl string
}

func ReadConfig() (TestConfig, error) {
	internalAPIUrl, found := os.LookupEnv(internalAPIUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", internalAPIUrlEnvName))
	}

	externalAPIUrl, found := os.LookupEnv(externalAPIUrlEnvName)
	if !found {
		return TestConfig{}, errors.New(fmt.Sprintf("failed to read %s environment variable", externalAPIUrlEnvName))
	}

	config := TestConfig{
		InternalAPIUrl: internalAPIUrl,
		ExternalAPIUrl: externalAPIUrl,
	}

	log.Printf("Read configuration: %+v", config)

	return config, nil
}
