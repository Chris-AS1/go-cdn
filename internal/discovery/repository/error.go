package repository

import "errors"

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceDisabled = errors.New("this discovery service is disabled")
