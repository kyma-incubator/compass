package tenantmapping

import (
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type StaticGroup struct {
	GroupName string   `json:"groupname"`
	Scopes    []string `json:"scopes"`
}

type StaticGroups []StaticGroup

//go:generate mockery -name=StaticGroupRepository -output=automock -outpkg=automock -case=underscore
type StaticGroupRepository interface {
	Get(groupnames []string) StaticGroups
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

func (r *staticGroupRepository) Get(groupnames []string) StaticGroups {
	result := []StaticGroup{}

	for _, groupname := range groupnames {
		if staticGroup, ok := r.data[groupname]; ok {
			result = append(result, staticGroup)
		} else {
			log.Warnf("static group with name %s not found", groupname)
		}
	}

	return result
}

// getGroupScopes get all scopes from group array, without duplicates
func (groups StaticGroups) GetGroupScopes() string {
	scopeMap := make(map[string]bool)
	filteredScopes := []string{}

	for _, group := range groups {
		for _, scope := range group.Scopes {
			_, ok := scopeMap[scope]
			if !ok {
				filteredScopes = append(filteredScopes, scope)
				scopeMap[scope] = true
			}
		}
	}

	return strings.Join(filteredScopes, " ")
}
