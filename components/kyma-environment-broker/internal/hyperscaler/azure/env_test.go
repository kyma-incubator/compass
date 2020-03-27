package azure

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
)

func Test_mapRegion(t *testing.T) {
	type args struct {
		hyperscalerType hyperscaler.HyperscalerType
		planID          string
		region          string
	}
	tests := []struct {
		name       string
		args       args
		wantRegion string
		wantErr    bool
	}{
		{
			name: "invalid gcp mapping",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.GcpPlanID,
				region:          "munich",
			},
			wantRegion: "",
			wantErr:    true,
		},
		{
			name: "valid gcp mapping",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.GcpPlanID,
				region:          "europe-west1",
			},
			wantRegion: "westeurope",
			wantErr:    false,
		},
		{
			name: "unknown planid",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          "microsoftcloud",
				region:          "",
			},
			wantRegion: "",
			wantErr:    true,
		},
		{
			name: "unknown hyperscaler",
			args: args{
				hyperscalerType: "microsoftcloud",
				planID:          broker.AzurePlanID,
				region:          "",
			},
			wantRegion: "",
			wantErr:    true,
		},
		{
			name: "invalid azure region",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.AzurePlanID,
				region:          "",
			},
			wantRegion: "",
			wantErr:    true,
		},
		{
			name: "valid azure region",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.AzurePlanID,
				region:          "westeurope",
			},
			wantRegion: "westeurope",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credentials := hyperscaler.Credentials{
				HyperscalerType: tt.args.hyperscalerType,
			}
			parameters := internal.ProvisioningParameters{
				PlanID: tt.args.planID,
				Parameters: internal.ProvisioningParametersDTO{
					Region: &tt.args.region,
				},
			}

			got, err := mapRegion(credentials, parameters)
			if tt.wantErr {
				require.NotNil(t, err, "mapRegion() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.Nil(t, err, "mapRegion() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, got, tt.wantRegion, "mapRegion() got = %v, wantRegion %v", got, tt.wantRegion)
		})
	}
}
