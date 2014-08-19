package main

import (
	//"database/sql"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	//"reflect"
)

import (
	//"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

func debugfmtArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	sep := ", "
	n := len(sep) * (len(args) - 1)
	for i := 0; i < len(args); i++ {
		n += 20
		n += (5 + (i / 10) + 1)
	}

	b := make([]byte, 0, n)
	buf := bytes.NewBuffer(b)
	buf.WriteRune('$')
	buf.WriteString(strconv.FormatInt(int64(1), 10))
	buf.WriteString(" = ")
	fmt.Fprintf(buf, "%v", args[0])
	for i, a := range args[1:] {
		buf.WriteString(sep)
		buf.WriteRune('$')
		buf.WriteString(strconv.FormatInt(int64(1+i+1), 10))
		buf.WriteString(" = ")
		fmt.Fprintf(buf, "%v", a)
	}
	return buf.String()
}

func RunQuery(sql string, args ...interface{}) (results []map[string]interface{}, err error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		debug_log("SQL: %s\nargs: %s\nERROR: %s\n", sql, debugfmtArgs(args...), err.Error())
		return nil, err
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

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
