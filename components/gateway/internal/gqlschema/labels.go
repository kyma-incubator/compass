package gqlschema

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Labels map[string]string

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
