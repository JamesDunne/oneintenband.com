package main

import (
	//"database/sql"
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	//"reflect"
)

import (
	//"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

// Pre-parse step to process HTML templates and add functions for templates to execute:
func uiTemplatesPreParse(tmpl *template.Template) *template.Template {
	log_info("Parsing HTML template files\n")
	return tmpl.Funcs(templateFunctions)
}

func flatten(query map[string][]string) (result map[string]string) {
	result = make(map[string]string)
	for key, values := range query {
		if len(values) == 0 {
			result[key] = ""
		} else {
			result[key] = values[0]
		}
	}
	return
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

	log_verbose("%s %s\n", req.Method, req.URL)

	// HACK(jsd): Temporary solution to serve static files.
	if staticPath, ok := web.MatchSimpleRouteRaw(req.URL.Path, "/static/"); ok {
		http.ServeFile(rsp, req, filepath.Join("../static/", staticPath))
		return nil
	} else if req.URL.Path == "/favicon.ico" {
		return web.NewError(nil, http.StatusNoContent, web.Empty)
	}

	// Parse URL route:
	route := strings.Split(req.URL.Path[1:], "/")
	log_verbose("route: %v\n", route)

	// Use first part of route as name of template to execute:
	templateName := strings.ToLower(route[0])
	if templateName == "" {
		templateName = "index"
	}
	log_verbose("templateName: '%s'\n", templateName)

	// Create a buffer to output the generated template to:
	bufWriter := bytes.NewBuffer(make([]byte, 0, 16384))

	// Execute the named template and output to the buffer:
	model := struct {
		Static   string
		Template string
		Route    []string
		Query    map[string]string
	}{
		Static:   staticHref,
		Template: templateName,
		Route:    route,
		// Flatten the query map of `[]string` values to `string` values:
		Query: flatten(req.URL.Query()),
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
