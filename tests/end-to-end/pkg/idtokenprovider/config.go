package idtokenprovider

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain        string `envconfig:"default=kyma.local"`
	UserEmail     string
	UserPassword  string
	ClientTimeout time.Duration `envconfig:"default=10s"` //Don't forget the unit!
}

type Config struct {
	DexConfig       dexConfig
	ClientConfig    clientConfig
	UserCredentials userCredentials
	RetryConfig     retryConfig}


type dexConfig struct {
	BaseUrl           string
	AuthorizeEndpoint string
	TokenEndpoint     string
}

type clientConfig struct {
	ID             string
	RedirectUri    string
	TimeoutSeconds time.Duration
}

type retryConfig struct {
	MaxAttempts uint
	Delay       time.Duration
}

type userCredentials struct {
	Username string
	Password string
}

func NewConfigFromEnv() (Config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	return initConfig(env), nil
}

func NewConfig(email, password, domain string, timeout time.Duration) (Config, error) {
	env := envConfig{Domain:domain,UserEmail:email,UserPassword:password, ClientTimeout:timeout}

	return initConfig(env), nil
}

func initConfig(env envConfig) Config{

	config := Config{
		DexConfig: dexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", env.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", env.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", env.Domain),
		},
		ClientConfig: clientConfig{
			ID:             "kyma-client",
			RedirectUri:    "http://127.0.0.1:5555/callback",
			TimeoutSeconds: env.ClientTimeout,
		},
		RetryConfig: retryConfig{
			MaxAttempts: 4,
			Delay:       3 * time.Second,
		},
	}

	config.UserCredentials = userCredentials{
		Username: env.UserEmail,
		Password: env.UserPassword,
	}
}

