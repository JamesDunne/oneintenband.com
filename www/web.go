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
