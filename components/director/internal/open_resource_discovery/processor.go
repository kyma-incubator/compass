package ord

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// ORDService missing godoc
//
//go:generate mockery --name=ORDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ORDService interface {
	ProcessApplication(ctx context.Context, appID string) error
	ProcessApplicationTemplate(ctx context.Context, appTemplateID string) error
	ProcessAppInAppTemplateContext(ctx context.Context, appTemplateID, appID string) error
}

// OperationsProcessor defines Open Resource Discovery operation processor
type OperationsProcessor struct {
	OrdSvc ORDService
}

// Process processes the given operation
func (p *OperationsProcessor) Process(ctx context.Context, operation *model.Operation) error {
	var opData OrdOperationData
	if err := json.Unmarshal(operation.Data, &opData); err != nil {
		return errors.Wrapf(err, "while unmarshalling operation with id %q", operation.ID)
	}

	if opData.ApplicationID != "" && opData.ApplicationTemplateID == "" {
		if err := p.OrdSvc.ProcessApplication(ctx, opData.ApplicationID); err != nil {
			return err
		}
	}

	// If there are AppID and AppTemplateID defined in the operation data - process app template static ord and process the app in te context of appTmpl
	if opData.ApplicationID != "" && opData.ApplicationTemplateID != "" {
		if err := p.OrdSvc.ProcessAppInAppTemplateContext(ctx, opData.ApplicationTemplateID, opData.ApplicationID); err != nil {
			return err
		}
	}

	// Aggregate only static ord
	if opData.ApplicationID == "" && opData.ApplicationTemplateID != "" {
		if err := p.OrdSvc.ProcessApplicationTemplate(ctx, opData.ApplicationTemplateID); err != nil {
			return err
		}
	}
	return nil
}
