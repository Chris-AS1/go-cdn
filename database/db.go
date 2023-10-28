package database

import (
	"database/sql"
	"fmt"
	"go-cdn/config"
	"go-cdn/consul"
	"log"
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
	log.Print("Connecting to Postgres")

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
	// Why
	if err.Error() == "no change" {
		return nil
	}
	return err
}

func (pg *PostgresClient) GetImageList(cfg *config.Config) (map[string]string, error) {
	con := pg.client
	// Variable Replacement of a table name not supported
	// rows, err := con.Query(fmt.Sprintf("SELECT * FROM %s", utils.EnvSettings.DatabaseTableName))
	// str := fmt.Sprintf("SELECT %s, %s FROM %s", cfg.DatabaseProvider.DatabaseColumnID, cfg.DatabaseProvider.DatabaseColumnFilename, cfg.DatabaseProvider.DatabaseName)
    var str string
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
