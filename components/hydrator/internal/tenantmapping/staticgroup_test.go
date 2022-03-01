package tenantmapping

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var validGroupsNameFileContent = `
- groupname: "admin"
  scopes:
  - "application:write"
  - "application:read"
- groupname: "developer"
  scopes:
  - "application:read"
`

var unknownGroupsNameFieldsFileContent = `
- groupname: "admin"
  notascope:
  - "application:write"
`

func TestStaticGroupRepository(t *testing.T) {
	t.Run("NewStaticGroupRepository returns repository instance populated from valid file", func(t *testing.T) {
		filePath := "static-groups.tmp.json"
		err := ioutil.WriteFile(filePath, []byte(validGroupsNameFileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticGroupRepository(filePath)

		require.NoError(t, err)
	})

	t.Run("NewStaticGroupRepository should fail when file does not exist", func(t *testing.T) {
		filePath := "not-existing.json"

		_, err := NewStaticGroupRepository(filePath)

		require.Equal(t, "while reading static groups file: open not-existing.json: no such file or directory", err.Error())
	})

	t.Run("NewStaticGroupRepository should fail when file is not valid", func(t *testing.T) {
		filePath := "static-groups.tmp.json"
		fileContent := `some-not-valid-content`
		err := ioutil.WriteFile(filePath, []byte(fileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticGroupRepository(filePath)

		require.Equal(t, "while unmarshalling static groups YAML: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type []tenantmapping.StaticGroup", err.Error())
	})

	t.Run("NewStaticGroupRepository should fail when file content contains unknown fields", func(t *testing.T) {
		filePath := "static-groups.tmp.json"
		err := ioutil.WriteFile(filePath, []byte(unknownGroupsNameFieldsFileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticGroupRepository(filePath)

		require.Equal(t, "while unmarshalling static groups YAML: error unmarshaling JSON: while decoding JSON: json: unknown field \"notascope\"", err.Error())
	})

	t.Run("returns single matching staticGroup", func(t *testing.T) {
		repo := staticGroupRepository{
			data: map[string]StaticGroup{
				"developer": StaticGroup{
					GroupName: "developer",
					Scopes:    []string{"application:read"},
				},
			},
		}
		groupnames := []string{"developer", "admin"}
		staticGroup := repo.Get(context.TODO(), groupnames)

		require.Equal(t, int(1), len(staticGroup))
		require.Equal(t, "developer", staticGroup[0].GroupName)
		require.Equal(t, []string{"application:read"}, staticGroup[0].Scopes)
	})

	t.Run("returns multiple matching staticGroup", func(t *testing.T) {
		repo := staticGroupRepository{
			data: map[string]StaticGroup{
				"developer": StaticGroup{
					GroupName: "developer",
					Scopes:    []string{"application:read"},
				},
				"admin": StaticGroup{
					GroupName: "admin",
					Scopes:    []string{"application:read", "application:write"},
				},
			},
		}
		groupnames := []string{"developer", "admin"}
		staticGroup := repo.Get(context.TODO(), groupnames)

		require.Equal(t, int(2), len(staticGroup))

		require.Equal(t, "developer", staticGroup[0].GroupName)
		require.Equal(t, []string{"application:read"}, staticGroup[0].Scopes)

		require.Equal(t, "admin", staticGroup[1].GroupName)
		require.Equal(t, []string{"application:read", "application:write"}, staticGroup[1].Scopes)
	})

	t.Run("returns empty array when staticGroup does not exist", func(t *testing.T) {
		repo := staticGroupRepository{}

		staticGroup := repo.Get(context.TODO(), []string{"non-existing"})

		require.Equal(t, int(0), len(staticGroup))
	})
}
