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
	const batchSize = 50
	targetDB = targetDB.Session(&gorm.Session{
		CreateBatchSize: batchSize,
	})

	batch := make([]*db.Author, 0, batchSize)
	for fileScanner.Scan() {
		authorLine := fileScanner.Text()
		if authorLine != "" {
			authorObj := db.Author{Name: authorLine}
			batch = append(batch, &authorObj)
		}
		linesNum++
		if len(batch) == batchSize {
			targetDB.Create(batch)
			batch = make([]*db.Author, 0, batchSize)
		}
	}
	targetDB.Create(batch)

	fmt.Printf("Finished importing: %v lines processed\n", linesNum)
}
