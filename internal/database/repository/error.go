package repository

import "errors"

var ErrDatabaseOp = errors.New("error on database operation")
var ErrKeyDoesNotExist = errors.New("key does not exist")
