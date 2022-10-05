package main

import (
	"database/sql"
	"fmt"
	"go-cdn/utils"
	"log"

	_ "github.com/lib/pq"
)

func dbConnection() *sql.DB {
	log.Print("Connecting to DB...")
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

func GetImageList() map[string]string {
	con := dbConnection()

	// Variable Replacement of a table name not supported
	// rows, err := con.Query(fmt.Sprintf("SELECT * FROM %s", utils.EnvSettings.DatabaseTableName))
	str := fmt.Sprintf("SELECT %s, %s FROM %s", utils.EnvSettings.DatabaseIDColumn, utils.EnvSettings.DatabaseFilenameColumn, utils.EnvSettings.DatabaseTableName)
	log.Print(str)
	// BUG
	// rows, err := con.Query(str, utils.EnvSettings.DatabaseIDColumn, utils.EnvSettings.DatabaseFilenameColumn)
	rows, err := con.Query(str)

	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	v := make(map[string]string)

	for rows.Next() {
		var i string
		var rea string

		err := rows.Scan(&i, &rea)
		if err != nil {
			log.Panic(err)
		}
		log.Print(i + " " + rea)
		v[i] = rea

	}
	log.Fatal(v)

	return v
}
