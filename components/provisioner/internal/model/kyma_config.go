package model

type KymaComponent string

type KymaConfig struct {
	ID                  string
	Release             Release
	Components          []KymaComponentConfig
	GlobalConfiguration Configuration
	ClusterID           string
}

type Release struct {
	Id            string
	Version       string
	TillerYAML    string
	InstallerYAML string
}

type GithubRelease struct {
	Id         int     `json:"id"`
	Name       string  `json:"name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	Url  string `json:"browser_download_url"`
}

type KymaComponentConfig struct {
	ID            string
	Component     KymaComponent
	Namespace     string
	Configuration Configuration
	KymaConfigID  string
}

type Configuration struct {
	ConfigEntries []ConfigEntry `json:"configEntries"`
}

type ConfigEntry struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

func NewConfigEntry(key, val string, secret bool) ConfigEntry {
	return ConfigEntry{
		Key:    key,
		Value:  val,
		Secret: secret,
	}
}
