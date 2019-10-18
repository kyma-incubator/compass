package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestSystemAuthGetReferenceObjectType(t *testing.T) {
	t.Run("GetReferenceObjectType returns ApplicationReference for SystemAuth referenced by the Application", func(t *testing.T) {
		appID := uuid.New()
		sysAuth := SystemAuth{
			AppID: str.Ptr(appID.String()),
		}

		refObjType, err := sysAuth.GetReferenceObjectType()

		require.NoError(t, err)
		require.Equal(t, ApplicationReference, refObjType)
	})

	t.Run("GetReferenceObjectType returns RuntimeReference for SystemAuth referenced by the Runtime", func(t *testing.T) {
		runtimeID := uuid.New()
		sysAuth := SystemAuth{
			RuntimeID: str.Ptr(runtimeID.String()),
		}

		refObjType, err := sysAuth.GetReferenceObjectType()

		require.NoError(t, err)
		require.Equal(t, RuntimeReference, refObjType)
	})

	t.Run("GetReferenceObjectType returns IntegrationSystemReference for SystemAuth referenced by the Integration System", func(t *testing.T) {
		intSysID := uuid.New()
		sysAuth := SystemAuth{
			IntegrationSystemID: str.Ptr(intSysID.String()),
		}

		refObjType, err := sysAuth.GetReferenceObjectType()

		require.NoError(t, err)
		require.Equal(t, IntegrationSystemReference, refObjType)
	})

	t.Run("GetReferenceObjectType returns error when called on SystemAuth with all reference properties set to nil", func(t *testing.T) {
		sysAuth := SystemAuth{}

		_, err := sysAuth.GetReferenceObjectType()

		require.EqualError(t, err, "unknown reference object type")
	})
}
