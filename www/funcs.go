package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

func Fetch(url string) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	debug_log("fetch: start GET %s\n", url)
	rsp, err := http.Get(url)
	if err != nil {
		debug_log("fetch: error GET %s ERROR %s\n", url, err.Error())
		return results, err
	}

	defer rsp.Body.Close()

	debug_log("fetch: %d GET %s\n", rsp.StatusCode, url)
	if rsp.StatusCode != 200 {
		return results, nil
	}

	err = json.NewDecoder(rsp.Body).Decode(&results)
	if err != nil {
		debug_log("fetch: %d GET %s json decode error %s\n", rsp.StatusCode, url, err.Error())
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
	"arr_new": func() []interface{} {
		return make([]interface{}, 0, 10)
	},
	"arr_append": func(a []interface{}, b interface{}) []interface{} {
		return append(a, b)
	},
	"arr_reverse": func(a []interface{}) []interface{} {
		reversed := make([]interface{}, len(a))
		for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = a[j], a[i]
		}
		return reversed
	},
	"time_parse": func(a string) (time.Time, error) {
		return time.Parse("2006-01-02T15:04:05-0700", a)
	},
	"time_now": func() time.Time {
		return time.Now()
	},
	"time_after": func(a time.Time, b time.Time) bool {
		return a.After(b)
	},
})
