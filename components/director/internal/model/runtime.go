package model

import (
	"fmt"
	"time"
)

type Runtime struct {
	ID          string
	Name        string
	Description *string
	Tenant      string
	Labels      map[string][]string
	Annotations map[string]string
	Status      *RuntimeStatus
	// directive for checking auth
	AgentAuth interface{}
}

type RuntimeStatus struct {
	Condition RuntimeStatusCondition
	Timestamp time.Time
}

type RuntimeStatusCondition string

const (
	RuntimeStatusConditionInitial RuntimeStatusCondition = "INITIAL"
	RuntimeStatusConditionReady   RuntimeStatusCondition = "READY"
	RuntimeStatusConditionFailed  RuntimeStatusCondition = "FAILED"
)

func (r *Runtime) AddLabel(key string, values []string) {
	if r.Labels == nil {
		r.Labels = make(map[string][]string)
	}

	if _, exists := r.Labels[key]; !exists {
		r.Labels[key] = r.uniqueStrings(values)
		return
	}

	r.Labels[key] = r.uniqueStrings(append(r.Labels[key], values...))
}

func (r *Runtime) DeleteLabel(key string, valuesToDelete []string) error {
	currentValues, exists := r.Labels[key]

	if !exists {
		return fmt.Errorf("label %s doesn't exist", key)
	}

	if len(valuesToDelete) == 0 {
		delete(r.Labels, key)
		return nil
	}

	set := r.mapFromSlice(currentValues)
	for _, val := range valuesToDelete {
		delete(set, val)
	}

	filteredValues := r.sliceFromMap(set)
	if len(filteredValues) == 0 {
		delete(r.Labels, key)
		return nil
	}

	r.Labels[key] = filteredValues
	return nil
}

func (r *Runtime) AddAnnotation(key string, value string) error {
	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}

	if _, exists := r.Annotations[key]; exists {
		return fmt.Errorf("annotation %s does already exist", key)
	}

	r.Annotations[key] = value
	return nil
}

func (r *Runtime) DeleteAnnotation(key string) error {
	if _, exists := r.Annotations[key]; !exists {
		return fmt.Errorf("annotation %s doesn't exist", key)
	}

	delete(r.Annotations, key)
	return nil
}

func (r *Runtime) uniqueStrings(in []string) []string {
	set := r.mapFromSlice(in)
	return r.sliceFromMap(set)
}

func (r *Runtime) mapFromSlice(in []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, i := range in {
		set[i] = struct{}{}
	}

	return set
}

func (r *Runtime) sliceFromMap(set map[string]struct{}) []string {
	var items []string
	for key := range set {
		items = append(items, key)
	}

	return items
}

type RuntimeInput struct {
	Name        string
	Description *string
	Labels      map[string][]string
	Annotations map[string]string
}

