package main

import (
	//"database/sql"
	"html/template"
	"net/http"
)

import (
	//"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

var templateFunctions template.FuncMap = template.FuncMap(map[string]interface{}{
	// 'Add' function to add two integers:
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"query": func(sql string) (rows []map[string]interface{}, err error) {
		return nil, nil
	},
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

	// Define some functions to call the database with from the templates:

	// Execute the named template:
	bufWriter := rsp
	err := uiTmpl.ExecuteTemplate(bufWriter, req.URL.Path, req)
	if web.AsWebError(err, http.StatusInternalServerError).RespondHTML(rsp) {
		return
	}

	return
}
