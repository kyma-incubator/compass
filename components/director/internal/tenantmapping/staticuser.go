package tenantmapping

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

type StaticUser struct {
	Username string   `json:"username"`
	Tenants  []string `json:"tenants"`
	Scopes   []string `json:"scopes"`
}

//go:generate mockery -name=StaticUserRepository -output=automock -outpkg=automock -case=underscore
type StaticUserRepository interface {
	Get(username string) (StaticUser, error)
}

type staticUserRepository struct {
	data map[string]StaticUser
}

func NewStaticUserRepository(srcPath string) (*staticUserRepository, error) {
	staticUsersBytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading static users file")
	}

	var staticUsers []StaticUser
	if err := yaml.UnmarshalStrict(staticUsersBytes, &staticUsers, yaml.DisallowUnknownFields); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling static users YAML")
	}

	data := make(map[string]StaticUser)

	for _, su := range staticUsers {
		data[su.Username] = su
	}

	return &staticUserRepository{
		data: data,
	}, nil
}

func (r *staticUserRepository) Get(username string) (StaticUser, error) {
	if staticUser, ok := r.data[username]; ok {
		return staticUser, nil
	}

	return StaticUser{}, fmt.Errorf("static user with name %s not found", username)
}
