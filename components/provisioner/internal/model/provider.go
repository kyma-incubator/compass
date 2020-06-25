package model

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

type ProviderConfiguration interface {
	ToShootTemplate(namespace string, accountId string, subAccountId string) (*gardener_types.Shoot, error)
}
