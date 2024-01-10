package storage

import (
	"database/sql"
	"dev_meets/internal/config"
	"fmt"
)

func initDbConnection(cnf *config.Config) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cnf.Postgresql.Host, cnf.Postgresql.Port, cnf.Postgresql.User, cnf.Postgresql.Password, cnf.Postgresql.DB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}
