package db

import (
	"database/sql"
	"fmt"
	"log"
)

// Подключение к БД
func NewDatabase(dbHost string, dbPort int, dbUser string, dbPassword string, dbBase string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbBase)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	log.Println("Database was successfully connected!")
	return db, nil
}
