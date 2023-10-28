package database

import (
	"database/sql"
	"fmt"
	"go-cdn/config"
	"go-cdn/consul"
	"log"

	_ "github.com/lib/pq"
)

type PostgresClient struct {
	client *sql.DB
}

func NewPostgresClient(csl *consul.ConsulClient, cfg *config.Config) (*PostgresClient, error) {
	// Discovers postgres from Consul
	address, port, err := csl.DiscoverService(cfg.DatabaseProvider.DatabaseHost)
	if err != nil {
		return nil, err
	}

	log.Print("Connecting to Postgres")
    connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DatabaseProvider.DatabaseUsername,
		cfg.DatabaseProvider.DatabasePassword,
		address,
		port,
		"database_name_todo",
		cfg.DatabaseProvider.DatabaseSSL,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	return &PostgresClient{db}, err
}

func (pg *PostgresClient) GetImageList(cfg *config.Config) (map[string]string, error) {
	con := pg.client
	// Variable Replacement of a table name not supported
	// rows, err := con.Query(fmt.Sprintf("SELECT * FROM %s", utils.EnvSettings.DatabaseTableName))
	str := fmt.Sprintf("SELECT %s, %s FROM %s", cfg.DatabaseProvider.DatabaseColumnID, cfg.DatabaseProvider.DatabaseColumnFilename, cfg.DatabaseProvider.DatabaseTableName)
	log.Print(str)

	// BUG - To check again
	// rows, err := con.Query(str, utils.EnvSettings.DatabaseIDColumn, utils.EnvSettings.DatabaseFilenameColumn)

	rows, err := con.Query(str)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	v := make(map[string]string)

	for rows.Next() {
		var i string
		var row_n string

		if err := rows.Scan(&i, &row_n); err != nil {
			return nil, err
		}

		log.Print(i + " " + row_n)
		v[i] = row_n

	}

	log.Print(v)
	return v, nil
}
