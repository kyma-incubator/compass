package azure

import (
	"testing"

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
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "invalid gcp mapping",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.GcpPlanID,
				region:          "munich",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "valid gcp mapping",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.GcpPlanID,
				region:          "europe-west1",
			},
			want:    "westeurope",
			wantErr: false,
		},
		{
			name: "unknown planid",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          "microsoftcloud",
				region:          "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unknown hyperscaler",
			args: args{
				hyperscalerType: "microsoftcloud",
				planID:          broker.AzurePlanID,
				region:          "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid azure region",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.AzurePlanID,
				region:          "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "valid azure region",
			args: args{
				hyperscalerType: hyperscaler.Azure,
				planID:          broker.AzurePlanID,
				region:          "westeurope",
			},
			want:    "westeurope",
			wantErr: false,
		},
		// TODO: Add test cases.
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
			if (err != nil) != tt.wantErr {
				t.Errorf("mapRegion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("mapRegion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
