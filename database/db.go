package database

import (
	"database/sql"
	"errors"
	"fmt"
	"go-cdn/config"
	"go-cdn/consul"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PostgresClient struct {
	client *sql.DB
}

func NewPostgresClient(csl *consul.ConsulClient, cfg *config.Config) (*PostgresClient, error) {
	pg_client := &PostgresClient{}
	connStr, err := pg_client.GetConnectionString(csl, cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	pg_client.client = db
	return pg_client, err
}

// Handle the termination of the connection
func (pg *PostgresClient) CloseConnection() error {
	err := pg.client.Close()
	return err
}

// Retrieves the connection string. Interrogates Consul if set
func (pg *PostgresClient) GetConnectionString(csl *consul.ConsulClient, cfg *config.Config) (string, error) {
	var err error
	var address string
	var port int

	if cfg.Consul.ConsulEnable {
		// Discovers postgres from Consul
		address, port, err = csl.DiscoverService(cfg.DatabaseProvider.DatabaseAddress)
		if err != nil {
			return "", err
		}
	} else {
		cfg_adr := strings.Split(cfg.DatabaseProvider.DatabaseAddress, ":")
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
func (pg *PostgresClient) MigrateDB() error {
	driver, err := postgres.WithInstance(pg.client, &postgres.Config{})
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
func (pg *PostgresClient) AddFile(id_hash string, filename string, content []byte) error {
	con := pg.client
	// hash := utils.RandStringBytes(6)
	_, err := con.Exec(`INSERT INTO fs_entities (id_hash, filename, content) VALUES ($1, $2, $3)`, id_hash, filename, content)
	return err
}

// Removes the file from teh database, if present
func (pg *PostgresClient) RemoveFile(id_hash string) error {
	con := pg.client
	_, err := con.Exec(`DELETE FROM fs_entities WHERE id_hash=$1`, id_hash)
	return err
}

// Queries the files saved on the database
func (pg *PostgresClient) GetFileList() (*map[string]int, error) {
	// Variable Replacement of a table name not supported
	// rows, err := con.Query(fmt.Sprintf("SELECT * FROM %s", utils.EnvSettings.DatabaseTableName))
	// str := fmt.Sprintf("SELECT %s, %s FROM %s", cfg.DatabaseProvider.DatabaseColumnID, cfg.DatabaseProvider.DatabaseColumnFilename, cfg.DatabaseProvider.DatabaseName)

	// BUG - To check again
	// rows, err := con.Query(str, utils.EnvSettings.DatabaseIDColumn, utils.EnvSettings.DatabaseFilenameColumn)

	con := pg.client

	rows, err := con.Query(fmt.Sprintf("SELECT id, id_hash, filename FROM %s", "fs_entities"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	available_files := make(map[string]int)

	for rows.Next() {
		var id int
		var id_hash string
		var filename string
		if err := rows.Scan(&id, &id_hash, &filename); err != nil {
			return nil, err
		}
		available_files[string(id_hash)] = id
	}

	return &available_files, nil
}

// Queries the specified file saved on the database
// TODO Return a structure representing the file
func (pg *PostgresClient) GetFile(id_hash string) ([]byte, error) {
	con := pg.client
	rows, err := con.Query(fmt.Sprintf("SELECT id, id_hash, filename, content FROM %s WHERE id_hash=%s", "fs_entities", id_hash))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []byte
	for rows.Next() {
		var id int
		var id_hash string
		var filename string
		if err := rows.Scan(&id, &id_hash, &filename, &content); err != nil {
			return nil, err
		}
	}

	return content, err
}
