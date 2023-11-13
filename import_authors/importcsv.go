package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lalathealter/olist/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	targetDB := db.Use()
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(originFile)
	fileScanner.Split(bufio.ScanLines)

	linesNum := 0
	rowsInserted := int64(0)
	const batchSize = 3
	targetDB = targetDB.Session(&gorm.Session{
		CreateBatchSize: batchSize,
	})
	targetDB.Statement.AddClause(clause.OnConflict{DoNothing: true})

	batch := make([]*db.Author, 0, batchSize)
	for fileScanner.Scan() {
		authorLine := fileScanner.Text()
		if authorLine != "" {
			authorObj := db.Author{Name: authorLine}
			batch = append(batch, &authorObj)
		}
		linesNum++
		if len(batch) == batchSize {
			dbc := targetDB.Create(batch)
			if dbc.Error != nil {
				panic(dbc.Error)
			}
			rowsInserted += dbc.RowsAffected
			batch = make([]*db.Author, 0, batchSize)
		}
	}

	dbc := targetDB.Create(batch)
	if dbc.Error != nil {
		panic(dbc.Error)
	}

	rowsInserted += dbc.RowsAffected

	fmt.Printf("Finished importing:\n%d lines processed\n", linesNum)
	fmt.Printf("%d rows inserted\n", rowsInserted)
}
