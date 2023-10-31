package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const TableAuthors = "authors"
const AuthorName = "name"

var InsertAuthorStmt = fmt.Sprintf(`
  INSERT INTO %s (%s) VALUES ($1)
  ON CONFLICT DO NOTHING`,
	TableAuthors, AuthorName,
)

var SelectAuthorsStmt = fmt.Sprintf(`
  SELECT *
  FROM %s
  WHERE %s LIKE $1
  `, TableAuthors, AuthorName,
)

var Instance *sql.DB

func init() {
	db, err := connect()
	if err != nil {
		log.Panic(err)
	}
	Instance = db
}

func connect() (*sql.DB, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		return nil, err
	}

	dbuser := os.Getenv("dbuser")
	dbpassword := os.Getenv("dbpassword")
	dbname := os.Getenv("dbname")
	dbhost := os.Getenv("dbhost")
	dbport := os.Getenv("dbport")

	connString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
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
