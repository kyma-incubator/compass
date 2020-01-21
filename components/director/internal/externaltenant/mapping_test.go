package externaltenant_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
)

func TestMapTenants(t *testing.T) {
	const (
		testProvider        = "testProvider"
		mappingOverrideName = "custom-name"
		mappingOverrideID   = "custom-id"
	)
	type args struct {
		srcPath          string
		provider         string
		mappingOverrides externaltenant.MappingOverrides
	}
	testCases := []struct {
		name    string
		args    args
		want    []externaltenant.TenantMappingInput
		wantErr bool
	}{
		{
			name: "should pass",
			args: args{
				srcPath:  "./testdata/valid_tenants.json",
				provider: testProvider,
				mappingOverrides: externaltenant.MappingOverrides{
					Name: mappingOverrideName,
					ID:   mappingOverrideID,
				},
			},
			want: []externaltenant.TenantMappingInput{
				{Name: "default", ExternalTenantID: "id-default", Provider: testProvider},
				{Name: "foo", ExternalTenantID: "id-foo", Provider: testProvider},
				{Name: "bar", ExternalTenantID: "id-bar", Provider: testProvider},
			},
			wantErr: false,
		},
		{
			name: "should fail during reading file",
			args: args{
				srcPath:          "invalid",
				provider:         "",
				mappingOverrides: externaltenant.MappingOverrides{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should fail during unmarshalling json",
			args: args{
				srcPath:          "./testdata/invalid.json",
				provider:         "",
				mappingOverrides: externaltenant.MappingOverrides{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := externaltenant.MapTenants(testCase.args.srcPath, testCase.args.provider, testCase.args.mappingOverrides)
			if (err != nil) != testCase.wantErr {
				t.Errorf("MapTenants() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			assert.Equal(t, testCase.want, got)
		})
	}
}
