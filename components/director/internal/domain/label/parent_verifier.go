package label

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

type ParentAccessVerifier struct {
	repo LabelRepository
}

func NewDefaultParentAccessVerifier() *ParentAccessVerifier {
	conv := NewConverter()
	return &ParentAccessVerifier{
		repo: NewRepository(conv),
	}
}

func NewParentAccessVerifier(labelRepo LabelRepository) *ParentAccessVerifier {
	return &ParentAccessVerifier{
		repo: labelRepo,
	}
}

func (p *ParentAccessVerifier) Verify(ctx context.Context, parentResourceType resource.Type, parentID string) error {
	tnt, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return err
	}

	labelableObject := resourceTypeToLabelableObject(parentResourceType)
	if labelableObject == "" {
		return errors.Errorf("unknown labelable object for resource %s", parentResourceType)
	}

	lbl, err := p.repo.GetByKeyGlobal(ctx, labelableObject, parentID, globalSubaccountIDLabelKey)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return errors.Errorf("the parent of type %s with id %s does not have %q label", parentResourceType, parentID, globalSubaccountIDLabelKey)
		}
		return errors.Wrapf(err, "cannot retrieve %q label for parent of type %s with id %s", globalSubaccountIDLabelKey, parentResourceType, parentID)
	}
	value, ok := lbl.Value.(string)
	if !ok {
		return errors.Errorf("unexpected type of %q label, expect: string, got: %T", globalSubaccountIDLabelKey, lbl.Value)
	}

	if value == tnt.ExternalID {
		return nil
	}

	return errors.Errorf("the provided tenant %s and the parent tenant %s do not match", tnt.ExternalID, value)
}

func resourceTypeToLabelableObject(r resource.Type) model.LabelableObject {
	if r == resource.ApplicationTemplate {
		return model.AppTemplateLabelableObject
	}
	return ""
}
