package main

import (
	//"database/sql"
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

import (
	//"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

func RunQuery(sql string) (results []map[string]interface{}, err error) {
	debug_log("query: %s\n", sql)
	rows, err := db.Query(sql)
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
			colMap[colNames[i]] = colValues[i]
		}
		results = append(results, colMap)
	}

	return results, nil
}

var templateFunctions template.FuncMap = template.FuncMap(map[string]interface{}{
	// 'Add' function to add two integers:
	"add":   func(a, b int) int { return a + b },
	"sub":   func(a, b int) int { return a - b },
	"query": RunQuery,
})

// Pre-parse step to process HTML templates and add functions for templates to execute:
func uiTemplatesPreParse(tmpl *template.Template) *template.Template {
	return tmpl.Funcs(templateFunctions)
}

// Handles all web requests:
func requestHandler(rsp http.ResponseWriter, req *http.Request) {
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
		return
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
		Static string
	}{
		Static: "/static",
	}
	err := uiTmpl.ExecuteTemplate(bufWriter, templateName, model)
	if web.AsWebError(err, http.StatusInternalServerError).RespondHTML(rsp) {
		return
	}

	// Write the buffer's contents to the HTTP response:
	_, err = bufWriter.WriteTo(rsp)
	if web.AsWebError(err, http.StatusInternalServerError).RespondHTML(rsp) {
		return
	}
	return
}
