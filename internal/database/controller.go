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
	AddFile(ctx context.Context, file *mod.StoredFile) error
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
	file, err := c.repo.GetFile(ctx, id_hash_search)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (c *Controller) GetFileList(ctx context.Context) (*[]mod.StoredFile, error) {
	l, err := c.repo.GetFileList(ctx)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (c *Controller) AddFile(ctx context.Context, file *mod.StoredFile) error {
	if err := c.repo.AddFile(ctx, file); err != nil {
		return err
	}
	return nil
}

func (c *Controller) RemoveFile(ctx context.Context, id_hash string) error {
	if err := c.repo.RemoveFile(ctx, id_hash); err != nil {
		return err
	}
	return nil
}

func (c *Controller) Close() error {
	return c.repo.CloseConnection()
}
