package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const TableAuthors = "authors"
const AuthorId = "id"
const AuthorName = "name"
const AuthorsPaginationLimit = 5

var InsertAuthorStmt = fmt.Sprintf(`
  INSERT INTO %s (%s) VALUES ($1)
  ON CONFLICT DO NOTHING`,
	TableAuthors, AuthorName,
)

var SelectAuthorsStmt = fmt.Sprintf(`
  SELECT name
  FROM %s
  WHERE LOWER(%s) LIKE (CONCAT('%%',LOWER($1::text),'%%'))
  LIMIT %d
  OFFSET %d * ($2 - 1)
  `, TableAuthors, AuthorName, AuthorsPaginationLimit, AuthorsPaginationLimit,
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

type BookModel struct {
	Name    string `json:"name"`
	Edition int    `json:"edition"`
	PubYear int    `json:"publication_year"`
	Authors []int  `json:"authors"`
}

var AuthorsExistStmt = fmt.Sprintf(`
  SELECT EVERY(EXISTS(
    SELECT * 
    FROM %s 
    WHERE %s = EL
  ))
  FROM UNNEST($1::int[]) EL;
  `, TableAuthors, AuthorId)

var ErrAuthorsOfBookDontExist = errors.New("Authors' ids of the book provided are not present in the database")

func SearchForAuthors(authorIds []int) error {
	db := Instance
	res := db.QueryRow(AuthorsExistStmt, pq.Array(authorIds))

	doExist := false
	err := res.Scan(&doExist)
	if err != nil {
		return err
	} else if !doExist {
		return ErrAuthorsOfBookDontExist
	}

	return nil
}

const (
	TableBooks  = "books"
	BookId      = "id"
	BookName    = "name"
	BookEdition = "edition"
	BookPubYear = "publication_year"
)

var BookInsertStmt = fmt.Sprintf(`
  INSERT INTO %s(%s, %s, %s)
  VALUES($1, $2, $3)
  RETURNING %s;
  `, TableBooks, BookName, BookEdition, BookPubYear, BookId)

const (
	TableBooksAuthorsLinks = "authors_books_links"
	BALinksBookId          = "book_id"
	BALinksAuthorId        = "author_id"
)

var BooksAuthorsLinkInsertStmt = fmt.Sprintf(`
  INSERT INTO %s(%s, %s)
  VALUES (%%d, $1)
  `, TableBooksAuthorsLinks, BALinksBookId, BALinksAuthorId)

func InsertBook(book BookModel) error {
	tx, err := Instance.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res := tx.QueryRow(BookInsertStmt, book.Name, book.Edition, book.PubYear)

	var newBookId int
	err = res.Scan(&newBookId)
	if err != nil {
		return err
	}

	fixedBookIdStmt := fmt.Sprintf(BooksAuthorsLinkInsertStmt, newBookId)
	stmt, err := tx.Prepare(fixedBookIdStmt)
	if err != nil {
		return err
	}
	for _, authorId := range book.Authors {
		stmt.Exec(authorId)
	}

	err = tx.Commit()
	return err
}
