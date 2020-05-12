package util

import (
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func StringPtr(str string) *string {
	return &str
}

func IntPtr(val int) *int {
	return &val
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

func TimePtr(time time.Time) *time.Time {
	return &time
}

func UnwrapStr(strPtr *string) string {
	if strPtr == nil {
		return ""
	}
	return *strPtr
}

func UnwrapInt(intPtr *int) int {
	if intPtr == nil {
		return 0
	}
	return *intPtr
}
