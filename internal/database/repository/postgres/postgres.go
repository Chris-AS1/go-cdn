package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/database/repository"
	"go-cdn/internal/discovery/controller"
	"go-cdn/internal/tracing"
	mod "go-cdn/pkg/model"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
)

type PostgresRepository struct {
	client *sql.DB
}

func New(ctx context.Context, dc *discovery.Controller, cfg *config.Config) (*PostgresRepository, error) {
	_, span := tracing.Tracer.Start(ctx, "pg/New")
	defer span.End()

	repo := &PostgresRepository{}
	conStr, err := repo.getConnectionString(dc, cfg)
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
func (r *PostgresRepository) CloseConnection() error {
	err := r.client.Close()
	return err
}

// Retrieves the connection string. Interrogates Consul if set
func (r *PostgresRepository) getConnectionString(dc *discovery.Controller, cfg *config.Config) (string, error) {
	address, err := dc.DiscoverService(cfg.Database.DatabaseAddress)
	if err != nil {
		return "", err
	}

	sslmode := ""
	switch cfg.Database.DatabaseSSL {
	case false:
		sslmode = "disable"
	case true:
		sslmode = "enable"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&connect_timeout=5",
		cfg.Database.DatabaseUsername,
		cfg.Database.DatabasePassword,
		address,
		cfg.Database.DatabaseName,
		sslmode,
	)

	return connStr, nil
}

// Apply all up-migrations under ./migrations
func (r *PostgresRepository) migrateDB() error {
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
func (r *PostgresRepository) AddFile(ctx context.Context, file *mod.StoredFile) error {
	_, span := tracing.Tracer.Start(ctx, "pg/AddFile")
	span.SetAttributes(attribute.String("pg.hash", file.IDHash),
		attribute.String("pg.filename", file.Filename))
	defer span.End()

	con := r.client
	// hash := utils.RandStringBytes(6)
	_, err := con.Exec(`INSERT INTO fs_entities (id_hash, filename, content) VALUES ($1, $2, $3)`, file.IDHash, file.Filename, file.Content)
	return err
}

// Removes the file from the database, if present
func (r *PostgresRepository) RemoveFile(ctx context.Context, id_hash string) error {
	_, span := tracing.Tracer.Start(ctx, "pg/RemoveFile")
	span.SetAttributes(attribute.String("pg.hash", id_hash))
	defer span.End()

	con := r.client
	_, err := con.Exec(`DELETE FROM fs_entities WHERE id_hash=$1`, id_hash)
	return err
}

// Queries the specified file saved on the database
func (r *PostgresRepository) GetFile(ctx context.Context, id_hash_search string) (*mod.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "pg/GetFile")
	span.SetAttributes(attribute.String("pg.hash", id_hash_search))
	defer span.End()

	con := r.client
	rows, err := con.Query("SELECT id, id_hash, filename, content FROM fs_entities WHERE id_hash=$1", id_hash_search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scanned := false
	var id int
	var id_hash string
	var filename string
	var content []byte
	for rows.Next() {
		if err := rows.Scan(&id, &id_hash, &filename, &content); err != nil {
			return nil, err
		}
		scanned = true
	}
	if !scanned {
		return nil, repository.ErrKeyDoesNotExist
	}

	return &mod.StoredFile{IDHash: id_hash, Filename: filename, Content: content}, err
}

// Retrieves a list of current files
func (r *PostgresRepository) GetFileList(ctx context.Context) (*[]mod.StoredFile, error) {
	_, span := tracing.Tracer.Start(ctx, "pg/GetFileList")
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
