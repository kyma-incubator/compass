package tenantmapping

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

type StaticGroup struct {
	GroupName string   `json:"groupname"`
	Tenants   []string `json:"tenants"`
	Scopes    []string `json:"scopes"`
}

//go:generate mockery -name=StaticUserRepository -output=automock -outpkg=automock -case=underscore
type StaticGroupRepository interface {
	Get(groupname string) (StaticGroup, error)
}

type staticGroupRepository struct {
	data map[string]StaticGroup
}

func NewStaticGroupRepository(srcPath string) (*staticGroupRepository, error) {
	staticGroupsBytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading static groups file")
	}

	var staticGroups []StaticGroup
	if err := yaml.UnmarshalStrict(staticGroupsBytes, &staticGroups, yaml.DisallowUnknownFields); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling static groups YAML")
	}

	data := make(map[string]StaticGroup)

	for _, sg := range staticGroups {
		data[sg.GroupName] = sg
	}

	return &staticGroupRepository{
		data: data,
	}, nil
}

func (r *staticUserRepository) Get(username string) (StaticUser, error) {
	if staticUser, ok := r.data[username]; ok {
		return staticUser, nil
	}

	return StaticUser{}, fmt.Errorf("static user with name %s not found", username)
}
