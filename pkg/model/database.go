package model

type DatabaseType string

const (
	DatabaseTypePostgres = DatabaseType("postgres")
)

type StoredFile struct {
	IDHash   string `json:"id_hash"`
	Filename string `json:"filename"`
	Content  []byte `json:"content,omitempty"`
}
