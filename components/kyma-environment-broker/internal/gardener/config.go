package gardener

type Config struct {
	Project        string `envconfig:"default=gardenerProject"`
	KubeconfigPath string `envconfig:"default=./dev/kubeconfig.yaml"`
}
