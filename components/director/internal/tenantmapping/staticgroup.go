package tenantmapping

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

// StaticGroup missing godoc
type StaticGroup struct {
	GroupName string   `json:"groupname"`
	Scopes    []string `json:"scopes"`
}

// StaticGroups missing godoc
type StaticGroups []StaticGroup

// StaticGroupRepository missing godoc
//go:generate mockery --name=StaticGroupRepository --output=automock --outpkg=automock --case=underscore
type StaticGroupRepository interface {
	Get(ctx context.Context, groupnames []string) StaticGroups
}

type staticGroupRepository struct {
	data map[string]StaticGroup
}

// NewStaticGroupRepository missing godoc
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

// Get missing godoc
func (r *staticGroupRepository) Get(ctx context.Context, groupnames []string) StaticGroups {
	result := []StaticGroup{}

	for _, groupname := range groupnames {
		if staticGroup, ok := r.data[groupname]; ok {
			result = append(result, staticGroup)
		} else {
			log.C(ctx).Warnf("Static group with name %s not found", groupname)
		}
	}

	return result
}

// GetGroupScopes get all scopes from group array, without duplicates
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
