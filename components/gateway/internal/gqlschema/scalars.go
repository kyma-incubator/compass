package gqlschema

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Labels map[string]string
type Tenant map[string]string

func (y *Tenant) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y Tenant) MarshalGQL(w io.Writer) {}

type Timestamp map[string]string

func (y *Timestamp) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y Timestamp) MarshalGQL(w io.Writer) {}

type Annotations map[string]string

func (y *Annotations) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y Annotations) MarshalGQL(w io.Writer) {}

type HttpHeaders map[string]string

func (y *HttpHeaders) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y HttpHeaders) MarshalGQL(w io.Writer) {}

type QueryParams map[string]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y QueryParams) MarshalGQL(w io.Writer) {}

type Clob map[string]string

func (y *Clob) UnmarshalGQL(v interface{}) error {
	return nil
}
func (y Clob) MarshalGQL(w io.Writer) {}

func (y *Labels) UnmarshalGQL(v interface{}) error {

	if v == nil {
		return nil
	}
	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.New("some error")
	}

	labels, err := y.convertToLabels(value)
	if err != nil {
		return errors.New("some error 2")
	}
	*y = labels

	return nil
}

//
func (y Labels) MarshalGQL(w io.Writer) {

	bytes, err := json.Marshal(y)
	if err != nil {
		fmt.Println("aaa")
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		fmt.Println("bbb")
		return
	}
}

//

func (y *Labels) convertToLabels(labels map[string]interface{}) (Labels, error) {
	result := make(map[string]string)
	for k, v := range labels {
		val, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("given value `%v` must be a string", v)
		}
		result[k] = val
	}
	return result, nil
}
