package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lalathealter/olist/db"
)

func main() {
	fmt.Println("Hello world!")

	mux := http.NewServeMux()

	mux.HandleFunc("/authors", func(w http.ResponseWriter, r *http.Request) {
		dbConn := db.Instance

		queryVals := r.URL.Query()
		namePattern := queryVals.Get("like")
		if namePattern == "" {
			namePattern = "%"
		}

		rows, err := dbConn.Query(db.SelectAuthorsStmt, namePattern)
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseSlice)
	})

	http.ListenAndServe("localhost:5000", mux)

}
