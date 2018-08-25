package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
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

func OpenFileJSON(path string) (results map[string]interface{}, err error) {
	results = make(map[string]interface{})

	var b []byte

	debug_log("openFileJSON: '%s'\n", path)
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("path cannot be absolute")
	}

	var absPath string
	absPath, err = filepath.Abs(filepath.Join(html_path(), path))
	if err != nil {
		return nil, fmt.Errorf("could not join html path with relative path")
	}

	debug_log("openFileJSON: readFile('%s')\n", absPath)
	b, err = ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &results)
	if err != nil {
		return nil, err
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

const fbTime = "2006-01-02T15:04:05-0700"

func parseFbTime(a string) (time.Time, error) {
	return time.Parse(fbTime, a)
}

var templateFunctions = template.FuncMap(map[string]interface{}{
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
		return "", fmt.Errorf("cannot stringify")
	},
	// URI-escape a string:
	"uri": func(a string) string {
		return url.QueryEscape(a)
	},
	"html": func(a string) template.HTML {
		return template.HTML(a)
	},
	"query":        RunQuery,
	"fetch":        Fetch,
	"openFileJSON": OpenFileJSON,
	"upcoming": func(a []interface{}) []interface{} {
		upcoming := make([]interface{}, 0, 10)
		for _, event := range a {
			if startTime, err := parseFbTime(event.(map[string]interface{})["start_time"].(string)); err == nil {
				if startTime.After(time.Now()) {
					upcoming = append([]interface{}{event}, upcoming...)
				}
			}
		}
		return upcoming
	},
	"past": func(a []interface{}) []interface{} {
		past := make([]interface{}, 0, 10)
		for _, event := range a {
			if startTime, err := parseFbTime(event.(map[string]interface{})["start_time"].(string)); err == nil {
				if !startTime.After(time.Now()) {
					past = append(past, event)
				}
			}
		}
		return past
	},
	"fb_time": parseFbTime,
	"month":   func(t time.Time) string { return t.Format("Jan") },
	"day":     func(t time.Time) string { return t.Format("02") },
	"time":    func(t time.Time) string { return t.Format("03:04 PM") },
})
