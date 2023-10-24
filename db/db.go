package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const envpath = "../.env"

const TableAuthors = "authors"
const AuthorName = "name"

func Connect() (*sql.DB, error) {
	godotenv.Load(envpath)
	dbuser := os.Getenv("dbuser")
	dbpassword := os.Getenv("dbpassword")
	dbname := os.Getenv("dbname")
	dbhost := os.Getenv("dbhost")
	dbportStr := os.Getenv("dbport")
	dbport, _ := strconv.Atoi(dbportStr)

	connString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		dbuser, dbpassword, dbname, dbhost, dbport)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
