package labeldef_test

import "errors"

func fixTenant() string {
	return "tenant"
}

func fixError() error {
	return errors.New("some error")
}
