package main

import (
	//"database/sql"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	//"reflect"
)

import (
	//"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

func RunQuery(sql string, args ...interface{}) (results []map[string]interface{}, err error) {
	debug_log("query: %s\n", sql)
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	debug_log("colnames: %s\n", colNames)

	results = make([]map[string]interface{}, 0, 10)

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
	"html": func(a string) template.HTML {
		return template.HTML(a)
	},
	"query": RunQuery,
})

// Pre-parse step to process HTML templates and add functions for templates to execute:
func uiTemplatesPreParse(tmpl *template.Template) *template.Template {
	return tmpl.Funcs(templateFunctions)
}

// Handles all web requests:
func requestHandler(rsp http.ResponseWriter, req *http.Request) (werr *web.Error) {
	// Set RemoteAddr for forwarded requests:
	{
		ip := req.Header.Get("X-Real-IP")
		if ip == "" {
			ip = req.Header.Get("X-Forwarded-For")
		}
		if ip != "" {
			req.RemoteAddr = ip
		}
	}

	verbose_log("%s %s\n", req.Method, req.URL)

	// HACK(jsd): Temporary solution to serve static files.
	if staticPath, ok := web.MatchSimpleRouteRaw(req.URL.Path, "/static/"); ok {
		http.ServeFile(rsp, req, filepath.Join("../static/", staticPath))
		return nil
	} else if req.URL.Path == "/favicon.ico" {
		return web.NewError(nil, http.StatusNoContent, web.Empty)
	}

	// Decide which template to execute:
	templateName := req.URL.Path[1:]
	if templateName == "" {
		templateName = "index"
	}

	verbose_log("templateName: '%s'\n", templateName)

	// Create a buffer to write a response to:
	bufWriter := bytes.NewBuffer(make([]byte, 0, 16384))

	// Execute the named template:
	model := struct {
		Static   string
		Template string
	}{
		Static:   "/static",
		Template: templateName,
	}
	err := uiTmpl.ExecuteTemplate(bufWriter, templateName, model)
	werr = web.AsErrorHTML(err, http.StatusInternalServerError)
	if werr != nil {
		return
	}

	// Write the buffer's contents to the HTTP response:
	_, err = bufWriter.WriteTo(rsp)
	werr = web.AsErrorHTML(err, http.StatusInternalServerError)
	if werr != nil {
		return
	}
	return
}
