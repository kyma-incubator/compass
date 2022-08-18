package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	applicationTemplatesDirectoryPath = "/data/templates/"
	integrationSystemsDirectoryPath   = "/data/int-systems/"
)

//go:generate mockery --name=appTmplService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type appTmplService interface {
	GetByNameAndRegion(ctx context.Context, name string, region interface{}) (*model.ApplicationTemplate, error)
	Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error)
}

//go:generate mockery --name=intSysRepo --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type intSysRepo interface {
	Create(ctx context.Context, item model.IntegrationSystem) error
	Exists(ctx context.Context, id string) (bool, error)
}

type DataLoader struct {
	transaction persistence.Transactioner
	appTmplSvc  appTmplService
	intSysRepo  intSysRepo
}

func NewDataLoader(tx persistence.Transactioner, appTmplSvc appTmplService, intSysRepo intSysRepo) *DataLoader {
	return &DataLoader{
		transaction: tx,
		appTmplSvc:  appTmplSvc,
		intSysRepo:  intSysRepo,
	}
}

func (d *DataLoader) LoadData(ctx context.Context, readDir func(dirname string) ([]fs.FileInfo, error), readFile func(filename string) ([]byte, error)) error {
	integrationSystems, err := d.loadIntegrationSystems(ctx, readDir, readFile)
	if err != nil {
		return errors.Wrap(err, "failed while loading integration systems")
	}

	appTemplateInputs, err := d.loadAppTemplates(ctx, readDir, readFile)
	if err != nil {
		return errors.Wrap(err, "failed while loading application templates")
	}

	tx, err := d.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "Error while beginning transaction")
	}
	defer d.transaction.RollbackUnlessCommitted(ctx, tx)
	ctxWithTx := persistence.SaveToContext(ctx, tx)

	if err = d.upsertIntegrationSystems(ctxWithTx, integrationSystems); err != nil {
		return errors.Wrap(err, "failed while upserting integration systems")
	}

	if err = d.upsertAppTemplates(ctxWithTx, appTemplateInputs); err != nil {
		return errors.Wrap(err, "failed while upserting application templates")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing transaction")
	}

	return nil
}

func (d *DataLoader) loadIntegrationSystems(ctx context.Context, readDir func(dirname string) ([]fs.FileInfo, error), readFile func(filename string) ([]byte, error)) ([]model.IntegrationSystem, error) {
	files, err := readDir(integrationSystemsDirectoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with integration systems files [%s]", integrationSystemsDirectoryPath)
	}

	var integrationSystems []model.IntegrationSystem
	for _, f := range files {
		log.C(ctx).Infof("Loading integration systems from file: %s", f.Name())

		if filepath.Ext(f.Name()) != ".json" {
			return nil, apperrors.NewInternalError(fmt.Sprintf("unsupported file format %q, supported format: json", filepath.Ext(f.Name())))
		}

		bytes, err := readFile(integrationSystemsDirectoryPath + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading integration systems file %q", integrationSystemsDirectoryPath+f.Name())
		}

		var integrationSystemsFromFile []model.IntegrationSystem
		if err := json.Unmarshal(bytes, &integrationSystemsFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling integration systems from file %s", integrationSystemsDirectoryPath+f.Name())
		}
		log.C(ctx).Infof("Successfully loaded integration systems from file: %s", f.Name())
		integrationSystems = append(integrationSystems, integrationSystemsFromFile...)
	}

	return integrationSystems, nil
}

func (d *DataLoader) loadAppTemplates(ctx context.Context, readDir func(dirname string) ([]fs.FileInfo, error), readFile func(filename string) ([]byte, error)) ([]model.ApplicationTemplateInput, error) {
	files, err := readDir(applicationTemplatesDirectoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with application templates files [%s]", applicationTemplatesDirectoryPath)
	}

	var appTemplateInputs []model.ApplicationTemplateInput
	for _, f := range files {
		log.C(ctx).Infof("Loading application templates from file: %s", f.Name())

		if filepath.Ext(f.Name()) != ".json" {
			return nil, apperrors.NewInternalError(fmt.Sprintf("unsupported file format %q, supported format: json", filepath.Ext(f.Name())))
		}

		bytes, err := readFile(applicationTemplatesDirectoryPath + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading application templates file %q", applicationTemplatesDirectoryPath+f.Name())
		}

		var templatesFromFile []model.ApplicationTemplateInput
		if err := json.Unmarshal(bytes, &templatesFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling application templates from file %s", applicationTemplatesDirectoryPath+f.Name())
		}
		log.C(ctx).Infof("Successfully loaded application templates from file: %s", f.Name())
		appTemplateInputs = append(appTemplateInputs, templatesFromFile...)
	}

	return appTemplateInputs, nil
}

func (d *DataLoader) upsertIntegrationSystems(ctx context.Context, intSystems []model.IntegrationSystem) error {
	for _, intSystem := range intSystems {
		log.C(ctx).Infof(fmt.Sprintf("Checking if integration system with id %q already exists", intSystem.ID))
		exist, err := d.intSysRepo.Exists(ctx, intSystem.ID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error while checking if integration system with id %q exists", intSystem.ID))
		}
		if !exist {
			log.C(ctx).Infof(fmt.Sprintf("Cannot find integration system with id %q. Creation triggered...", intSystem.ID))
			if err = d.intSysRepo.Create(ctx, intSystem); err != nil {
				return errors.Wrap(err, fmt.Sprintf("error while creating integration system with id %q", intSystem.ID))
			}
			log.C(ctx).Infof(fmt.Sprintf("Successfully registered integration system with id %q", intSystem.ID))
		} else {
			log.C(ctx).Infof(fmt.Sprintf("Integration system with id %q already exists", intSystem.ID))
		}
	}

	return nil
}

func (d *DataLoader) upsertAppTemplates(ctx context.Context, appTemplateInputs []model.ApplicationTemplateInput) error {
	for _, appTmplInput := range appTemplateInputs {
		var region interface{}
		region, err := retrieveRegion(appTmplInput.Labels)
		if err != nil {
			return err
		}
		if region == "" {
			region = nil
		}

		log.C(ctx).Infof(fmt.Sprintf("Retrieving application template with name %q and region %s", appTmplInput.Name, region))
		_, err = d.appTmplSvc.GetByNameAndRegion(ctx, appTmplInput.Name, region)
		if err != nil {
			if !strings.Contains(err.Error(), "Object not found") {
				return errors.Wrap(err, fmt.Sprintf("error while getting application template with name %q and region %s", appTmplInput.Name, region))
			}

			log.C(ctx).Infof(fmt.Sprintf("Cannot find application template with name %q and region %s. Creation triggered...", appTmplInput.Name, region))
			templateID, err := d.appTmplSvc.Create(ctx, appTmplInput)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error while creating application template with name %q and region %s", appTmplInput.Name, region))
			}
			log.C(ctx).Infof(fmt.Sprintf("Successfully registered application template with id: %q", templateID))
		}
	}

	return nil
}

func retrieveRegion(labels map[string]interface{}) (string, error) {
	if labels == nil {
		return "", nil
	}

	region, exists := labels[tenant.RegionLabelKey]
	if !exists {
		return "", nil
	}

	regionValue, ok := region.(string)
	if !ok {
		return "", fmt.Errorf("%q label value must be string", tenant.RegionLabelKey)
	}
	return regionValue, nil
}
