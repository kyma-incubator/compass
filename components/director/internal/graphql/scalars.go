package graphql

import (
	"encoding/json"
	"io"
	"log"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

type Timestamp time.Time

func (y *Timestamp) UnmarshalGQL(v interface{}) error {

	if v == nil {
		return nil
	}

	tmpStr, ok := v.(string)

	if !ok {
		return errors.Errorf("unexpected input type: %T, should be string", v)

	}

	t, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return err
	}

	*y = Timestamp(t)

	return nil
}
func (y Timestamp) MarshalGQL(w io.Writer) {
	err := writeResponse(time.Time(y), w)
	if err != nil {
		log.Print(err)
		return
	}
}

type Tenant string

func (y *Tenant) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = Tenant(val)

	return nil
}

func (y Tenant) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["tenant"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %v", reflect.TypeOf(y))
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error with writing %v", reflect.TypeOf(y))
	}
}

type Labels map[string][]string

func (y *Labels) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}

	labels, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}
	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {
	bytes, err := json.Marshal(y)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Print(err)
		return
	}
}

type Annotations map[string]interface{}

func (y *Annotations) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}
	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected Annotations type: %T, should be map[string]interface{}", v)
	}

	*y = value

	return nil
}
func (y Annotations) MarshalGQL(w io.Writer) {
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}

type HttpHeaders map[string][]string

func (y *HttpHeaders) UnmarshalGQL(v interface{}) error {
	headers, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = headers

	return nil
}
func (y HttpHeaders) MarshalGQL(w io.Writer) {
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}

type QueryParams map[string][]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {

	params, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = params

	return nil
}
func (y QueryParams) MarshalGQL(w io.Writer) {
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}

type CLOB string

func (y *CLOB) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = CLOB(val)

	return nil
}
func (y CLOB) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["CLOB"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %v", reflect.TypeOf(y))
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error with writing %v", reflect.TypeOf(y))
	}
}

type PageCursor string

func (y *PageCursor) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = PageCursor(val)

	return nil
}
func (y PageCursor) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["pageCursor"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %v", reflect.TypeOf(y))
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error with writing %v", reflect.TypeOf(y))
	}
}

func writeResponse(in interface{}, w io.Writer) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return errors.Errorf("error with marshalling %v", reflect.TypeOf(in))
	}

	_, err = w.Write(bytes)
	if err != nil {
		return errors.Errorf("Error with writing %v", reflect.TypeOf(in))
	}
	return nil
}

func convertToString(in interface{}) (string, error) {
	if in == nil {
		return "", errors.New("input should not be nil")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.Errorf("unexpected input type: %T, should be map[string][]string", in)
	}

	return value, nil
}

func convertToMapStringStringArray(in interface{}) (map[string][]string, error) {
	result := make(map[string][]string)

	value, ok := in.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("unexpected input type: %T, should be map[string][]string", in)
	}

	for k, v := range value {
		val, ok := v.([]string)
		if !ok {
			return nil, errors.Errorf("given value `%v` must be a string array", v)
		}
		result[k] = val
	}
	return result, nil
}
