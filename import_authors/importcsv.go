package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lalathealter/olist/db"
)

var ErrNoOriginFile = errors.New("ERROR: no origin file was provided")

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	originFileName := flag.String("f", "", "file to import from")
	flag.Parse()
	fmt.Println(*originFileName)

	if *originFileName == "" {
		panic(ErrNoOriginFile)
	}

	originFile, err := os.Open(*originFileName)
	if err != nil {
		panic(err)
	}
	defer originFile.Close()

	targetDB, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer targetDB.Close()

	tx, err := targetDB.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare(db.InsertAuthorStmt)
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(originFile)
	fileScanner.Split(bufio.ScanLines)
	linesNum := 0
	for fileScanner.Scan() {
		authorLine := fileScanner.Text()
		if authorLine != "" {
			stmt.Exec(authorLine)
		}
		linesNum++
	}

	err = stmt.Close()
	if err != nil {
		panic(err)
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Finished importing: %v lines processed\n", linesNum)
}
