package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
)

func Fetch(url string) (results []map[string]interface{}, err error) {
	results = make([]map[string]interface{}, 0, 10)

	rsp, err := http.Get(url)
	if err != nil {
		return results, err
	}

	if rsp.StatusCode != 200 {
		return results, nil
	}

	err = json.NewDecoder(rsp.Body).Decode(results)
	if err != nil {
		return results, err
	}

	return results, nil
}

func RunQuery(sql string, args ...interface{}) (results []map[string]interface{}, err error) {
	results = make([]map[string]interface{}, 0, 10)

	rows, err := db.Query(sql, args...)
	if err != nil {
		debug_log("SQL: %s\nargs: %s\nERROR: %s\n", sql, debugfmtArgs(args...), err.Error())
		if failsafe {
			return results, nil
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		debug_log("SQL: %s\nargs: %s\nERROR: %s\n", sql, debugfmtArgs(args...), err.Error())
		if failsafe {
			return results, nil
		} else {
			return nil, err
		}
	}

	colValues := make([]interface{}, len(colNames))
	cols := make([]interface{}, len(colNames))
	for i := 0; i < len(colNames); i++ {
		cols[i] = &colValues[i]
	}
	for rows.Next() {
		err = rows.Scan(cols...)
		if err != nil {
			return nil, err
		}

		colMap := make(map[string]interface{})
		for i := 0; i < len(colNames); i++ {
			var rv interface{}
			v := colValues[i]
			switch r := v.(type) {
			case []byte:
				// Convert `[]byte` to `string`:
				rv = string(r)
			case nil:
				rv = ""
			default:
				rv = v
			}

			colMap[colNames[i]] = rv
		}
		results = append(results, colMap)
	}

	debug_log("SQL: %s\nargs: %s\ncolumns: %s\nrows: %d\n", sql, debugfmtArgs(args...), colNames, len(results))
	return results, nil
}

var templateFunctions template.FuncMap = template.FuncMap(map[string]interface{}{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"string": func(a interface{}) (string, error) {
		switch s := a.(type) {
		case string:
			return s, nil
		case []byte:
			return string(s), nil
		default:
			debug_log("string: %v\n", s)
		}
		if s, ok := a.(fmt.Stringer); ok {
			return s.String(), nil
		}
		return "", fmt.Errorf("Cannot stringify!")
	},
	// URI-escape a string:
	"uri": func(a string) string {
		return url.QueryEscape(a)
	},
	"html": func(a string) template.HTML {
		return template.HTML(a)
	},
	"query": RunQuery,
	"fetch": Fetch,
})
