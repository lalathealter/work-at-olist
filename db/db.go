package db

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var bindUse = func(db *gorm.DB) func() *gorm.DB {
	return func() *gorm.DB {
		return db
	}
}

var Use func() *gorm.DB

func init() {
	db, err := setup()
	if err != nil {
		panic(err)
	}
	Use = bindUse(db)
}

func setup() (*gorm.DB, error) {
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
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	db.AutoMigrate(
		&Book{}, &Author{}, &BookAuthorLink{},
	)
	return db, nil
}

type Book struct {
	ID      uint
	Name    string `gorm:"not null;index:book,unique"`
	Edition uint   `gorm:"not null;index:book,unique"`
	PubYear int    `gorm:"not null;index:book,unique"`
}

type BookAuthorLink struct {
	ID       uint
	BookID   uint `gorm:"not null;index:balink,unique"`
	Book     Book
	AuthorID uint `gorm:"not null;index:balink,unique"`
	Author   Author
}

type Author struct {
	ID   uint
	Name string `gorm:"not null;unique"`
}
