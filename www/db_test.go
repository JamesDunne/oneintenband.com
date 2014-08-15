package main

import (
	"database/sql"
	"testing"
	//"time"
)

func Test_query(t *testing.T) {
	var err error

	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		t.Fatal(err)
	}

	results, err := RunQuery("select Date, Contents from News order by Date desc limit 3")
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range results {
		t.Log(row)
	}
}
