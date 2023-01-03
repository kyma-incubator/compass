package formationconstraint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"text/template"
)

// FormationConstraintService represents the Formation Constraint service layer
//go:generate mockery --name=FormationConstraintService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationConstraintService interface {
	ListMatchingConstraints(ctx context.Context, formationTemplateID string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error)
}

type MatchingDetails struct {
	resourceType    model.ResourceType
	resourceSubtype string
}

type JoinPointDetails interface {
	GetMatchingDetails() MatchingDetails
}

type CRUDFormationOperationDetails struct {
	FormationType string
	FormationName string
	CheckScope    model.OperatorScopeType
	TenantID      string
}

func (d *CRUDFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    model.FormationResourceType,
		resourceSubtype: d.FormationType,
	}
}

type AssignFormationOperationDetails struct {
	ResourceType    model.ResourceType
	ResourceSubtype string
	ResourceID      string
	FormationType   string
	FormationID     string
	TenantID        string
}

func (d *AssignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    d.ResourceType,
		resourceSubtype: d.ResourceSubtype,
	}
}

type UnassignFormationOperationDetails struct {
	ResourceType    model.ResourceType
	ResourceSubtype string
	ResourceID      string
	FormationType   string
	FormationID     string
	TenantID        string
}

func (d *UnassignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    d.ResourceType,
		resourceSubtype: d.ResourceSubtype,
	}
}

type GenerateNotificationOperationDetails struct {
	ResourceID    string
	Assignment    model.FormationAssignment
	TargetSubtype string
}

func (d *GenerateNotificationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    model.ResourceType(d.Assignment.TargetType),
		resourceSubtype: d.TargetSubtype,
	}
}

type OperatorName string

type OperatorInput interface{}

type OperatorFunc func(input OperatorInput) (bool, error)

type OperatorInputConstructor func() OperatorInput

type JoinPointLocation struct {
	OperationName  model.TargetOperation
	ConstraintType model.FormationConstraintType
}

type ConstraintEngine struct {
	constraintSvc             FormationConstraintService
	operators                 map[OperatorName]OperatorFunc
	operatorInputConstructors map[OperatorName]OperatorInputConstructor
}

func (e *ConstraintEngine) EnforceConstraints(ctx context.Context, location JoinPointLocation, details JoinPointDetails, formationTemplateID string) (bool, error) {
	constraints, err := e.constraintSvc.ListMatchingConstraints(ctx, formationTemplateID, location, details.GetMatchingDetails())
	if err != nil {
		return false, err
	}

	var errs *multierror.Error
	result := true
	for _, mc := range constraints {
		operator, ok := e.operators[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q not found", mc.Operator),
			})
			result = false
			continue
		}

		operatorInputConstructor, ok := e.operatorInputConstructors[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator input constructor for operator %q not found", mc.Operator),
			})
			result = false
			continue
		}

		operatorInput := operatorInputConstructor()

		if err := parseTemplate(mc.InputTemplate, details, operatorInput); err != nil {
			log.C(ctx).Errorf("An error occured while parsing input template for formation constraint %q: %s", mc.Name, err.Error())
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Failed to parse operator input template for operator %q", mc.Operator),
			})
			result = false
			continue
		}

		operatorResult, err := operator(operatorInput)
		if err != nil {
			return false, errors.Wrapf(err, "An error occured while executing operator %q for formation constraint %q", mc.Operator, mc.Name)
		}

		if !operatorResult {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q is not satisfied", mc.Operator),
			})
		}

		result = result && operatorResult
	}

	return result, nil
}

func parseTemplate(tmpl string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return err
	}
	if err = json.Unmarshal(res.Bytes(), dest); err != nil {
		return err
	}

	if validatable, ok := dest.(inputvalidation.Validatable); ok {
		return validatable.Validate()
	}

	return nil
}
