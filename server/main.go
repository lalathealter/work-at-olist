package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lalathealter/olist/db"
)

func HandleMethodNotSupported(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint doesn't support the provided method",
		http.StatusNotImplemented)
}

type MethodMap map[string]http.HandlerFunc

func HandleEndpointMethods(mm *MethodMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, ok := (*mm)[r.Method]
		if !ok {
			HandleMethodNotSupported(w, r)
			return
		}

		handler(w, r)
	}
}

func main() {
	fmt.Println("Hello world!")

	mux := http.NewServeMux()

	authorsHandlers := &MethodMap{
		"GET": HandleGetAuthors,
	}
	mux.HandleFunc("/authors", HandleEndpointMethods(authorsHandlers))

	booksHandlers := &MethodMap{
		"POST": HandlePostBooks,
	}
	mux.HandleFunc("/books", HandleEndpointMethods(booksHandlers))
	http.ListenAndServe("localhost:5000", mux)

}

func HandlePostBooks(w http.ResponseWriter, r *http.Request) {
	bookObj := db.BookModel{}
	err := json.NewDecoder(r.Body).Decode(&bookObj)
	if err != nil {
		panic(err)
	}

	if bookObj.Name == "" {
		http.Error(w, "Book must have a non-empty name",
			http.StatusBadRequest)
		return
	}

	if len(bookObj.Authors) < 1 {
		http.Error(w, "Book must have at least one author",
			http.StatusBadRequest)
		return
	}

	err = db.SearchForAuthors(bookObj.Authors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.InsertBook(bookObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
