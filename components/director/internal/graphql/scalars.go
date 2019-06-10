package graphql

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

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
	err := writeResponse(y,w)
	if err != nil {
		log.Print(err)
		return
	}
}

type Timestamp time.Time

func (y *Timestamp) UnmarshalGQL(v interface{}) error {

	if v == nil {
		return nil
	}

	tmpStr, ok := v.(string)

	if !ok {
		return errors.New("Error with unmarshalling time") //TODO
	}

	t, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return errors.New("Error with unmarshalling time") //TODO
	}

	*y = Timestamp(t)

	return nil
}
func (y Timestamp) MarshalGQL(w io.Writer) {
	err := writeResponse(time.Time(y),w)
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

	val, ok := v.([]byte)
	if !ok {
		return errors.New("error with assertion")
	}

	var result map[string]interface{}

	err := json.Unmarshal([]byte(v), &result)
	if err != nil {
		return errors.Errorf("Error with unmarshaling: %v", err)
	}

	log.Print(result)

	*y = result

	return nil
}
func (y Annotations) MarshalGQL(w io.Writer) {
	err := writeResponse(y,w)
	if err != nil {
		log.Print(err)
		return
	}
}

type HttpHeaders map[string][]string

func (y *HttpHeaders) UnmarshalGQL(v interface{}) error {
	headers, err := convertToMapStringArrString(v)
	if err != nil{
		return err
	}

	*y = headers

	return nil
}
func (y HttpHeaders) MarshalGQL(w io.Writer) {
	err := writeResponse(y,w)
	if err != nil {
		log.Print(err)
		return
	}
}

type QueryParams map[string][]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {

	params, err := convertToMapStringArrString(v)
	if err != nil{
		return err
	}

	*y = params

	return nil
}
func (y QueryParams) MarshalGQL(w io.Writer) {
	err := writeResponse(y,w)
	if err != nil {
		log.Print(err)
		return
	}
}

type CLOB map[string]string

func (y *CLOB) UnmarshalGQL(v interface{}) error {
	return nil
	//TODO: impl
}
func (y CLOB) MarshalGQL(w io.Writer) {
	//TODO: impl
}

type Labels map[string][]string

func (y *Labels) UnmarshalGQL(v interface{}) error {

	labels, err := convertToMapStringArrString(v)
	if err != nil{
		return errors.New("abc")
	}

	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {

	bytes, err := json.Marshal(y)
	if err != nil {
		fmt.Println("label err 3")
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		fmt.Println("label err 4")
		return
	}
}

func (y *Labels) convertToLabels(labels map[string]interface{}) (Labels, error) {
	result := make(map[string]string)
	for k, v := range labels {
		val, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("given value `%v` must be a string", v)
		}
		result[k] = val
	}
	return nil, nil
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
	err := writeResponse(y,w)
	if err != nil {
		log.Print(err)
		return
	}
}

func writeResponse(in interface{}, w io.Writer) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return errors.Errorf("Error with marshalling %v", reflect.TypeOf(in))
	}

	_, err = w.Write(bytes)
	if err != nil {
		return errors.Errorf("Error with writing %v", reflect.TypeOf(in))
	}
	return nil
}

func convertToString(in interface{}) (string,error){
	if in == nil {
		return "", errors.New("error")
	}

	value, ok := in.(string)
	if !ok {
		return "",errors.New("Error with unmarshalling Tenant")
	}

	return value,nil
}

func convertToMapStringArrString(in interface{}) (map[string][]string,error) {
	if in == nil {
		return nil, errors.New("Error: input is nil")
	}

	val, ok := in.(string)
	if !ok {
		return nil,errors.New("Error: input is not a string")
	}

	var result map[string][]string

	err := json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil,errors.Errorf("Error with converting string to map[string][]string: %v", err)
	}

	return result,nil
}
