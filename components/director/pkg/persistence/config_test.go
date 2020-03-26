package persistence

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConnString(t *testing.T) {
	t.Run("should generate database connection string based on the configuration", func(t *testing.T) {
		expectedConnStr := "host=dbhost port=12345 user=dbuser password=dbpass dbname=dbname sslmode=enable"
		dbCfg := DatabaseConfig{
			User:     "dbuser",
			Password: "dbpass",
			Host:     "dbhost",
			Port:     "12345",
			Name:     "dbname",
			SSLMode:  "enable",
		}

		connStr := dbCfg.GetConnString()

		require.Equal(t, expectedConnStr, connStr)
	})
}
