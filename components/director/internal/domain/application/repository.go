package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

const labelTableName string = `"public"."labels"`
const scenarioKey string = "SCENARIOS"

type inMemoryRepository struct {
	store map[string]*model.Application
}

func NewRepository() *inMemoryRepository {
	return &inMemoryRepository{store: make(map[string]*model.Application)}
}

func (r *inMemoryRepository) GetByID(ctx context.Context, tenant, id string) (*model.Application, error) {
	application := r.store[id]

	if application == nil || application.Tenant != tenant {
		return nil, errors.New("application not found")
	}

	return application, nil
}

//TODO: remvoe this function after migrating to Database
func (r *inMemoryRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	application := r.store[id]

	if application == nil || application.Tenant != tenant {
		return false, nil
	}

	return true, nil
}

// TODO: Make filtering and paging
func (r *inMemoryRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	var items []*model.Application
	for _, item := range r.store {
		if item.Tenant == tenant {
			items = append(items, item)
		}
	}

	return &model.ApplicationPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

//TODO: @dbadura call database for apps id in label table
// TODO: @dbadura add pagination when PR-181 is merged
func (r *inMemoryRepository) ListByScenariosFromRuntime(ctx context.Context, tenantID string, runtimeID string, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT VALUE FROM %s WHERE TENANT_ID=$1 AND RUNTIME_ID=$2 AND KEY='SCENARIOS'`, labelTableName)

	var scenariosBinary interface{}
	err = persist.Get(&scenariosBinary, stmt, tenantID, runtimeID)

	if scenariosBinary == nil {
		return nil, errors.New("Runtime scenarios not found")
	}
	scenarios := getScenariosValues(scenariosBinary)

	var scenarioFilers []*labelfilter.LabelFilter

	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenarioFilers = append(scenarioFilers, &labelfilter.LabelFilter{Key: scenarioKey, Query: &query})
	}

	//TODO: change tenantID from UUID to String
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("tenant_ID is not parseable")
	}

	stmt, err = label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenantUUID, scenarioFilers)
	if err != nil {
		return nil, errors.Wrap(err, "while creating filter query")
	}

	var apps []interface{}

	err = persist.Select(&apps, stmt)

	if err != nil {
		if err == sql.ErrNoRows {
			return &model.ApplicationPage{
				Data:       nil,
				TotalCount: 0,
				PageInfo: &pagination.Page{
					StartCursor: "",
					EndCursor:   "",
					HasNextPage: false,
				},
			}, nil
		}
		return nil, errors.Wrap(err, "while getting application for runtime from DB")
	}

	var items []*model.Application

	for _, id := range apps {
		appID, ok := id.(string)
		if !ok {
			return nil, errors.New("while parsing application IDs")
		}

		app, found := r.store[appID]
		if found {
			items = append(items, app)
		}
	}

	return &model.ApplicationPage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
}

func (r *inMemoryRepository) Create(ctx context.Context, item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	found := r.findApplicationNameWithinTenant(item.Tenant, item.Name)
	if found {
		return errors.New("Application name is not unique within tenant")
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Update(ctx context.Context, item *model.Application) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	oldApplication := r.store[item.ID]
	if oldApplication == nil {
		return errors.New("application not found")
	}

	if oldApplication.Name != item.Name {
		found := r.findApplicationNameWithinTenant(item.Tenant, item.Name)
		if found {
			return errors.New("Application name is not unique within tenant")
		}
	}

	r.store[item.ID] = item

	return nil
}

func (r *inMemoryRepository) Delete(ctx context.Context, item *model.Application) error {
	if item == nil {
		return nil
	}

	delete(r.store, item.ID)

	return nil
}

func (r *inMemoryRepository) findApplicationNameWithinTenant(tenant, name string) bool {
	for _, app := range r.store {
		if app.Name == name && app.Tenant == tenant {
			return true
		}
	}
	return false
}

func getScenariosValues(scenariosJSON interface{}) []string {
	var scenarios []string

	scen, ok := scenariosJSON.([]byte)
	if !ok {
		return scenarios
	}

	err := json.Unmarshal([]byte(scen), &scenarios)
	if err != nil {
		return scenarios
	}

	return scenarios
}
