package main

import (
	"database/sql"
	"fmt"
	"go-cdn/utils"
	"log"

	_ "github.com/lib/pq"
)

func dbConnection() *sql.DB {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.EnvSettings.DatabaseUsername,
		utils.EnvSettings.DatabasePassword,
		utils.EnvSettings.DatabaseURL,
		utils.EnvSettings.DatabasePort,
		utils.EnvSettings.DatabaseName,
		utils.EnvSettings.DatabaseSSL)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func GetImageList() []string {
	log.Print("Connecting to DB...")
	con := dbConnection()

	// Variable Replacement of a table name not supported
	rows, err := con.Query(fmt.Sprintf(`SELECT $1 FROM %s`, utils.EnvSettings.DatabaseTableName),
		utils.EnvSettings.DatabaseFilenameColumn)

	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	var v []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			log.Fatal(err)
		}
		log.Print(r)
		v = append(v, r)
	}

	return v
}
