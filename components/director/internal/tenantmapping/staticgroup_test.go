package tenantmapping

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

var validGroupsNameFileContent string = `
- groupname: "admin"
  scopes:
  - "application:write"
  - "application:read"
  tenants: 
  - "3f2d9157-a0ba-4ff3-9d50-f5d6f9730eed"
- groupname: "developer"
  scopes:
  - "application:read"
  tenants: 
  - "555d9157-a1bf-4aa2-9a22-f5d6f9730aaf"
`

var unknownGroupsNameFieldsFileContent string = `
- groupname: "admin"
  notascope:
  - "application:write"
  tenants: 
  - "3f2d9157-a0ba-4ff3-9d50-f5d6f9730eed"
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

	t.Run("returns staticGroup instance when exists", func(t *testing.T) {
		tenantIDs := []string{uuid.New().String()}
		repo := staticGroupRepository{
			data: map[string]StaticGroup{
				"developer": StaticGroup{
					GroupName: "developer",
					Scopes:    []string{"application:read"},
					Tenants:   tenantIDs,
				},
			},
		}
		groupnames := []string{"developer", "admin"}
		staticGroup := repo.Get(groupnames)

		require.Equal(t, int(1), len(staticGroup))
		require.Equal(t, "developer", staticGroup[0].GroupName)
		require.Equal(t, tenantIDs, staticGroup[0].Tenants)
		require.Equal(t, []string{"application:read"}, staticGroup[0].Scopes)
	})

	t.Run("returns empty array when staticGroup does not exist", func(t *testing.T) {
		repo := staticGroupRepository{}

		staticGroup := repo.Get([]string{"non-existing"})

		require.Equal(t, int(0), len(staticGroup))
	})
}
