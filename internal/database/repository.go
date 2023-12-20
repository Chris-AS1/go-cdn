package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/consul"
	"go-cdn/internal/tracing"
	mod "go-cdn/pkg/model"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
)

type Repository struct {
	client *sql.DB
}

func NewPostgresRepo(csl *consul.ConsulClient, cfg *config.Config) (*Repository, error) {
	repo := &Repository{}
	conStr, err := repo.GetConnectionString(csl, cfg)
	if err != nil {
		return nil, err
	}

	con, err := sql.Open(string(mod.DatabaseTypePostgres), conStr)
	if err != nil {
		return nil, err
	}

	err = con.Ping()
	repo.client = con
	if err != nil {
		return nil, err
	}
	err = repo.migrateDB()
	return repo, err
}

// Handle the termination of the connection
func (r *Repository) CloseConnection() error {
	err := r.client.Close()
	return err
}

// Retrieves the connection string. Interrogates Consul if set
func (r *Repository) GetConnectionString(csl *consul.ConsulClient, cfg *config.Config) (string, error) {
	var err error
	var address string
	var port int

	if cfg.Consul.ConsulEnable {
		// Discovers Postgres from Consul
		address, port, err = csl.DiscoverService(cfg.DatabaseProvider.DatabaseAddress)
		if err != nil {
			return "", err
		}
	} else {
		cfg_adr := strings.Split(cfg.DatabaseProvider.DatabaseAddress, ":")
		if len(cfg_adr) != 2 {
			return "", fmt.Errorf("wrong address format")
		}
		address = cfg_adr[0]
		port, _ = strconv.Atoi(cfg_adr[1])
	}

	sslmode := ""
	switch cfg.DatabaseProvider.DatabaseSSL {
	case false:
		sslmode = "disable"
	case true:
		sslmode = "enable"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=5",
		cfg.DatabaseProvider.DatabaseUsername,
		cfg.DatabaseProvider.DatabasePassword,
		address,
		port,
		cfg.DatabaseProvider.DatabaseName,
		sslmode,
	)

	return connStr, nil
}

// Apply all up-migrations under ./migrations
func (r *Repository) migrateDB() error {
	driver, err := postgres.WithInstance(r.client, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // Relative, equal to ./migrations
		"postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}

// Adds the byte stream as file in the database
func (r *Repository) AddFile(ctx context.Context, id_hash string, filename string, content []byte) error {
	_, span := tracing.Tracer.Start(ctx, "pgAddFile")
	span.SetAttributes(attribute.String("pg.hash", id_hash),
		attribute.String("pg.filename", filename))
	defer span.End()

	con := r.client
	// hash := utils.RandStringBytes(6)
	_, err := con.Exec(`INSERT INTO fs_entities (id_hash, filename, content) VALUES ($1, $2, $3)`, id_hash, filename, content)
	return err
}

// Removes the file from the database, if present
func (r *Repository) RemoveFile(ctx context.Context, id_hash string) error {
	_, span := tracing.Tracer.Start(ctx, "pgAddFile")
	span.SetAttributes(attribute.String("pg.hash", id_hash))
	defer span.End()

	con := r.client
	_, err := con.Exec(`DELETE FROM fs_entities WHERE id_hash=$1`, id_hash)
	return err
}

// Queries the specified file saved on the database
func (r *Repository) GetFile(ctx context.Context, id_hash_search string) (*mod.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "pgGetFile")
	span.SetAttributes(attribute.String("pg.hash", id_hash_search))
	defer span.End()

	con := r.client
	rows, err := con.Query("SELECT id, id_hash, filename, content FROM fs_entities WHERE id_hash=$1", id_hash_search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int
	var id_hash string
	var filename string
	var content []byte
	for rows.Next() {
		if err := rows.Scan(&id, &id_hash, &filename, &content); err != nil {
			return nil, err
		}
	}

	return &mod.StoredFile{id_hash, filename, content}, err
}

// Retrieves a list of current files
func (r *Repository) GetFileList(ctx context.Context) (*[]mod.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "pgGetFileList")
	defer span.End()

	con := r.client
	rows, err := con.Query("SELECT id_hash, filename FROM fs_entities")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id_hash string
	var filename string
	file_list := []mod.StoredFile{}
	for rows.Next() {
		if err := rows.Scan(&id_hash, &filename); err != nil {
			return nil, err
		}

		file_list = append(file_list, mod.StoredFile{
			IDHash:   id_hash,
			Filename: filename,
			Content:  nil,
		})
	}

	return &file_list, err
}
