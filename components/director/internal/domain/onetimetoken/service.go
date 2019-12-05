package onetimetoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const requestForRuntime = `
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`

const requestForApplication = `
		mutation { generateApplicationToken (appID:"%s")
		  {
			token
		  }
		}`

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface {
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
}

type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

type service struct {
	cli          GraphQLClient
	connectorURL string
	sysAuthSvc   SystemAuthService
	appSvc       ApplicationService
	labelSvc     LabelRepository
}

func NewTokenService(gcli GraphQLClient, sysAuthSvc SystemAuthService, appSvc ApplicationService, labelService LabelRepository, connectorURL string) *service {
	return &service{cli: gcli, connectorURL: connectorURL, sysAuthSvc: sysAuthSvc, labelSvc: labelService, appSvc: appSvc}
}

type PairingIntegrationToken struct {
	IntegrationToken string `json:"integrationToken"`
}

func (s service) GenerateOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (model.OneTimeToken, error) {
	if tokenType == model.ApplicationReference {
		tnt, err := tenant.LoadFromContext(ctx)
		if err != nil {
			return model.OneTimeToken{}, errors.Wrap(err, "while getting tenant from context")
		}
		appTypeLabel, err := s.labelSvc.GetByKey(ctx, tnt, model.ApplicationLabelableObject, id, "applicationType")

		switch {
		case apperrors.IsNotFoundError(err):
			// do nothing
		case err != nil:
			return model.OneTimeToken{}, errors.Wrap(err, "while getting label `applicationType`")
		default:

			appType, ok := appTypeLabel.Value.(string)
			if !ok {
				return model.OneTimeToken{}, errors.New("while casting applicationType to string")
			}

			if appType != "SFSF" && appType != "S4HanaCloud" {
				break
			}

			app, err := s.appSvc.Get(ctx, id)
			if err != nil {
				return model.OneTimeToken{}, errors.New("while getting the application")
			}

			params := &url.Values{}
			params.Set("app", id)
			params.Set("name", app.Name)
			params.Set("type", appType)

			baseUrl, err := url.Parse("http://localhost:8082")
			if err != nil {
				return model.OneTimeToken{}, errors.Wrap(err, "while parsing sidecar url")
			}
			baseUrl.RawQuery = params.Encode()
			resp, err := http.Get(baseUrl.String())
			if err != nil {
				return model.OneTimeToken{}, errors.Wrap(err, "while making get request")
			}

			if resp.StatusCode != 200 {
				return model.OneTimeToken{}, fmt.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
			}
			if resp.Body == nil {
				return model.OneTimeToken{}, errors.New("empty response body")
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return model.OneTimeToken{}, errors.Wrap(err, "while reading resp body")
			}

			paringToken := PairingIntegrationToken{}
			err = json.Unmarshal(b, &paringToken)
			if err != nil {
				return model.OneTimeToken{}, errors.Wrap(err, "while unmarshalling JSON")
			}

			return model.OneTimeToken{Token: paringToken.IntegrationToken, ConnectorURL: "not applicable"}, nil

		}

	}

	sysAuthID, err := s.sysAuthSvc.Create(ctx, tokenType, id, nil)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrap(err, "while creating System Auth")
	}

	token, err := s.getOneTimeToken(ctx, sysAuthID, tokenType)
	if err != nil {
		return model.OneTimeToken{}, errors.Wrapf(err, "while generating onetime token for %s", tokenType)
	}

	return model.OneTimeToken{Token: token, ConnectorURL: s.connectorURL}, nil
}

func (s service) getOneTimeToken(ctx context.Context, id string, tokenType model.SystemAuthReferenceObjectType) (string, error) {
	var request *gcli.Request

	switch tokenType {
	case model.RuntimeReference:
		request = gcli.NewRequest(fmt.Sprintf(requestForRuntime, id))
	case model.ApplicationReference:
		request = gcli.NewRequest(fmt.Sprintf(requestForApplication, id))
	default:
		return "", errors.Errorf("cannot generate token for %T", tokenType)
	}

	output := ConnectorTokenModel{}
	err := s.cli.Run(ctx, request, &output)
	if err != nil {
		return "", errors.Wrapf(err, "while calling connector for %s one time token", tokenType)
	}

	return output.Token(tokenType), err
}
