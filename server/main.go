package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lalathealter/olist/db"
)

func main() {
	fmt.Println("Hello world!")

	mux := http.NewServeMux()

	mux.HandleFunc("/authors", HandleGetAuthors)

	http.ListenAndServe("localhost:5000", mux)

}

type PaginatedArray[T any] struct {
	Data         []T
	PageNum      int
	NextPageLink string
}

const pageQuery = "page"
const likeQuery = "like"

func HandleGetAuthors(w http.ResponseWriter, r *http.Request) {
	dbConn := db.Instance

	queryVals := r.URL.Query()
	namePattern := queryVals.Get(likeQuery)
	if namePattern == "" {
		namePattern = "%"
	}
	pageArg := queryVals.Get(pageQuery)
	pageNum, err := strconv.Atoi(pageArg)
	if err != nil || pageNum < 1 {
		pageNum = 1
		queryVals.Set(pageQuery, "1")
	}

	rows, err := dbConn.Query(db.SelectAuthorsStmt, namePattern, pageNum)
	if err != nil {
		panic(err.Error())
	}

	responseSlice := make([]string, 0)
	for rows.Next() {
		var name string
		err := rows.Scan(&name)

		if err != nil {
			panic(err)
		}
		responseSlice = append(responseSlice, name)
	}

	pagRes := PaginatedArray[string]{
		responseSlice, pageNum, "",
	}
	if len(responseSlice) >= db.AuthorsPaginationLimit {
		if err != nil {
			panic(err)
		}

		queryVals.Set(pageQuery, strconv.Itoa(pageNum+1))
		r.URL.RawQuery = queryVals.Encode()
		pagRes.NextPageLink = r.Host + r.URL.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pagRes)
}
