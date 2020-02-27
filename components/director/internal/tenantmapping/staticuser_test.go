package tenantmapping

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

var validUserNameFileContent string = `
- username: "admin"
  scopes:
  - "application:write"
  tenants: 
  - "3f2d9157-a0ba-4ff3-9d50-f5d6f9730eed"
- username: "developer"
  scopes:
  - "application:read"
  tenants: 
  - "555d9157-a1bf-4aa2-9a22-f5d6f9730aaf"
`

var unknownUserNameFieldsFileContent string = `
- username: "admin"
  scope:
  - "application:write"
  tenants: 
  - "3f2d9157-a0ba-4ff3-9d50-f5d6f9730eed"
`

func TestStaticUserRepository(t *testing.T) {
	t.Run("NewStaticUserRepository returns repository instance populated from valid file", func(t *testing.T) {
		filePath := "static-users.tmp.json"
		err := ioutil.WriteFile(filePath, []byte(validUserNameFileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticUserRepository(filePath)

		require.NoError(t, err)
	})

	t.Run("NewStaticUserRepository should fail when file does not exist", func(t *testing.T) {
		filePath := "not-existing.json"

		_, err := NewStaticUserRepository(filePath)

		require.Equal(t, "while reading static users file: open not-existing.json: no such file or directory", err.Error())
	})

	t.Run("NewStaticUserRepository should fail when file is not valid", func(t *testing.T) {
		filePath := "static-users.tmp.json"
		fileContent := `some-not-valid-content`
		err := ioutil.WriteFile(filePath, []byte(fileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticUserRepository(filePath)

		require.Equal(t, "while unmarshalling static users YAML: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type []tenantmapping.StaticUser", err.Error())
	})

	t.Run("NewStaticUserRepository should fail when file content contains unknown fields", func(t *testing.T) {
		filePath := "static-users.tmp.json"
		err := ioutil.WriteFile(filePath, []byte(unknownUserNameFieldsFileContent), 0644)
		require.NoError(t, err)
		defer func(t *testing.T) {
			err := os.Remove(filePath)
			require.NoError(t, err)
		}(t)

		_, err = NewStaticUserRepository(filePath)

		require.Equal(t, "while unmarshalling static users YAML: error unmarshaling JSON: while decoding JSON: json: unknown field \"scope\"", err.Error())
	})

	t.Run("returns StaticUser instance when exists", func(t *testing.T) {
		tenantIDs := []string{uuid.New().String()}
		repo := staticUserRepository{
			data: map[string]StaticUser{
				"admin@domain.local": StaticUser{
					Username: "admin@domain.local",
					Scopes:   []string{"application:read"},
					Tenants:  tenantIDs,
				},
			},
		}

		staticUser, err := repo.Get("admin@domain.local")

		require.NoError(t, err)
		require.Equal(t, "admin@domain.local", staticUser.Username)
		require.Equal(t, tenantIDs, staticUser.Tenants)
		require.Equal(t, []string{"application:read"}, staticUser.Scopes)
	})

	t.Run("returns error when StaticUser does not exist", func(t *testing.T) {
		repo := staticUserRepository{}

		_, err := repo.Get("non-existing@domain.local")

		require.Equal(t, "static user with name non-existing@domain.local not found", err.Error())
	})
}
