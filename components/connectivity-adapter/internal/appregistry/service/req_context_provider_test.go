package service

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
)

func Test_requestContextProvider_ForRequest(t *testing.T) {
	type fields struct {
		graphqlizer       *graphqlizer.Graphqlizer
		gqlFieldsProvider *graphqlizer.GqlFieldsProvider
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    RequestContext
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &requestContextProvider{
				graphqlizer:       tt.fields.graphqlizer,
				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
			}
			got, err := s.ForRequest(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ForRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ForRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
