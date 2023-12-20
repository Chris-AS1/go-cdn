package database

import (
	"context"
	"errors"
	mod "go-cdn/pkg/model"
)

var ErrDatabaseOp = errors.New("error on database operation")

type databaseRepository interface {
	GetFile(ctx context.Context, id_hash_search string) (*mod.StoredFile, error)
	GetFileList(ctx context.Context) (*[]mod.StoredFile, error)
	AddFile(ctx context.Context, id_hash string, filename string, content []byte) error
	RemoveFile(ctx context.Context, id_hash string) error
	CloseConnection() error
}

type Controller struct {
	repo databaseRepository
}

func NewController(repo databaseRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) GetFile(ctx context.Context, id_hash_search string) (*mod.StoredFile, error) {
	content, err := c.repo.GetFile(ctx, id_hash_search)
	if err != nil {
		return nil, ErrDatabaseOp
	}
	return content, nil
}

func (c *Controller) GetFileList(ctx context.Context) (*[]mod.StoredFile, error) {
	l, err := c.repo.GetFileList(ctx)
	if err != nil {
		return nil, ErrDatabaseOp
	}
	return l, nil
}

func (c *Controller) AddFile(ctx context.Context, id_hash string, filename string, content []byte) error {
	if err := c.repo.AddFile(ctx, id_hash, filename, content); err != nil {
		return ErrDatabaseOp
	}
	return nil
}

func (c *Controller) RemoveFile(ctx context.Context, id_hash string) error {
	if err := c.repo.RemoveFile(ctx, id_hash); err != nil {
		return ErrDatabaseOp
	}
	return nil
}

func (c *Controller) Close() error {
	return c.repo.CloseConnection()
}
