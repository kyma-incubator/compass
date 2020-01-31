package util

import "k8s.io/apimachinery/pkg/util/intstr"

func StringPtr(str string) *string {
	return &str
}

func BoolPtr(b bool) *bool {
	return &b
}

func BoolFromPtr(val *bool) bool {
	if val == nil {
		return false
	}

	return *val
}

func IntOrStrPtr(intOrStr intstr.IntOrString) *intstr.IntOrString {
	return &intOrStr
}
