package main

import (
	"database/sql"
	"testing"
	//"time"
)

func Test_query(t *testing.T) {
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("select to_char(date, 'YYYY-MM-DD') Date, Contents from News order by Date desc limit 3")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	t.Log(colNames)

	colBytes := make([]string, len(colNames))
	colValues := make([]interface{}, len(colNames))
	for rows.Next() {
		for i := 0; i < len(colNames); i++ {
			colValues[i] = &colBytes[i]
		}
		err = rows.Scan(colValues...)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(colBytes)

		//var date time.Time
		//var contents string
		//rows.Scan(&date, &contents)
		//t.Log(date.Format("2006-01-02"), contents)
	}
}
