package tenant

import (
	"time"

	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/config"
)

func getTestConfig(url string) config.TenantInfo {
	return config.TenantInfo{
		Endpoint:           url,
		RequestTimeout:     10 * time.Second,
		InsecureSkipVerify: true,
	}
}

func getTestConfigWithBrokenURL() config.TenantInfo {
	return config.TenantInfo{
		Endpoint:           "asdsf",
		RequestTimeout:     100,
		InsecureSkipVerify: true,
	}
}
