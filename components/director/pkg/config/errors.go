package config

import "github.com/pkg/errors"

var ValueNotFoundError = errors.New("value under specified path not found in configuration")
