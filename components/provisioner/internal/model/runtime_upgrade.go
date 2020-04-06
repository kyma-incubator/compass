package model

type UpgradeState string

const (
	UpgradeInProgress UpgradeState = "IN_PROGRESS"
	UpgradeSucceeded  UpgradeState = "SUCCEEDED"
	UpgradeFailed     UpgradeState = "FAILED"
	UpgradeRolledBack UpgradeState = "ROLLED_BACK"
)

type RuntimeUpgrade struct {
	Id                      string
	State                   UpgradeState
	ClusterId               string
	PreUpgradeKymaConfigId  string
	PostUpgradeKymaConfigId string

	PreUpgradeKymaConfig  KymaConfig `db:"-"`
	PostUpgradeKymaConfig KymaConfig `db:"-"`
}
