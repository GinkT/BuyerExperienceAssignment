package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Подключение к БД
func NewDatabase() (*sql.DB, error) {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"));
	if err != nil {
		log.Println("[Database] Invalid DB_PORT in config!")
		return nil, err
	}
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), dbPort, os.Getenv("DB_BASE"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	log.Println("[Database] Database was successfully connected.")
	return db, nil
}
