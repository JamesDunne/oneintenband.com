package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
)

import _ "github.com/lib/pq"

import (
	"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

const (
	dbConnectionString = "host=localhost sslmode=disable user=www password=band dbname=oneintenband "
)

var (
	base_folder = ""
	verbose     = false
	db          *sql.DB
)

func html_path() string { return base_folder + "/html" }

func verbose_log(fmt string, args ...interface{}) {
	if !verbose {
		return
	}
	log.Printf(fmt, args...)
}

var uiTmpl *template.Template

func main() {
	// Define our commandline flags:
	fs := flag.String("fs", ".", "Root directory of served files and templates")
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available are tcp, unix)")
	flag.BoolVar(&verbose, "v", true, "verbose logging")
	flag.Parse()

	// Parse all the URIs:
	listen_addr, err := base.ParseListenable(*fl_listen_uri)
	base.PanicIf(err)

	// Make directories we need:
	base_folder = base.CanonicalPath(path.Clean(*fs))

	// Open database:
	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Println(err)
		return
	}

	// Watch the html templates for changes and reload them:
	_, cleanup, err := web.WatchTemplates("ui", html_path(), "*.html", uiTemplatesPreParse, &uiTmpl)
	if err != nil {
		log.Println(err)
		return
	}
	defer cleanup()

	// Start the server:
	_, err = base.ServeMain(listen_addr, func(l net.Listener) error {
		return http.Serve(l, http.HandlerFunc(requestHandler))
	})
	if err != nil {
		log.Println(err)
		return
	}
}
