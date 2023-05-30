package handler

import (
	"net/http"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/machinebox/graphql"
)

// AdapterHandler processes received requests
type AdapterHandler struct {
	GqlClient *graphql.Client
}

// HandlerFunc is the implementation of AdapterHandler
func (a AdapterHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Infof("Received request")
	qglReq := graphql.NewRequest(`mutation {
  result: registerRuntime(
    in: {
      name: "runtime-create-update-delete"
      description: "runtime-1-description"
      labels: { ggg: ["hhh"] }
    }
  ) {
    id
    name
    description
    labels
}
}
`)
	qglReq.Header.Add("x-request-id", "YOOOOOOOOOOOOOO")
	qglReq.Header.Add("Tenant", "a1ee8d85-f71b-475d-adee-e8515141cf84")
	var formationTemplate directorSchema.FormationTemplate
	err := a.GqlClient.Run(ctx, qglReq, formationTemplate)
	if err != nil {
		log.C(ctx).Error(err.Error())
	}
}
