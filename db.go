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

func ReadRows() {
	log.Print("Connecting to DB...")
	con := dbConnection()

	// TODO
	rows, err := con.Query("SELECT * FROM TBD")
	defer rows.Close()

	if err != nil {
		log.Panic(err)
	}

	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			log.Fatal(err)
		}
		log.Print(v)
	}
}
