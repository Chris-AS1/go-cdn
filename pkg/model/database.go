package model

import "errors"

type DatabaseType string

var ErrKeyDoesNotExist = errors.New("key does not exist")

const (
	DatabaseTypePostgres = DatabaseType("postgres")
)

type StoredFile struct {
	IDHash   string `json:"id_hash"`
	Filename string `json:"filename"`
	Content  []byte `json:"content,omitempty"`
}
