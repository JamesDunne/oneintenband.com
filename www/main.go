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
	dbConnectionString = "host=localhost sslmode=disable dbname=band user=band password=band"
)

var (
	html_folder = ""
	staticHref  = ""
	verbose     = false
	debug       = false
	db          *sql.DB
)

func html_path() string { return base.CanonicalSymlinkPath(path.Clean(html_folder)) }

func log_info(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func log_verbose(fmt string, args ...interface{}) {
	if !verbose {
		return
	}
	log.Printf(fmt, args...)
}

func debug_log(fmt string, args ...interface{}) {
	if !debug {
		return
	}
	log.Printf(fmt, args...)
}

var uiTmpl *template.Template

func main() {
	// Define our commandline flags:
	flag.StringVar(&html_folder, "html", "./html", "Directory of HTML template files")
	flag.StringVar(&staticHref, "static", "static", "HREF prefix to static files")
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available are tcp, unix)")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&debug, "d", false, "debug logging")
	flag.Parse()

	// Parse all the URIs:
	listen_addr, err := base.ParseListenable(*fl_listen_uri)
	base.PanicIf(err)

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
		return http.Serve(l, web.ReportErrors(web.Log(web.DefaultErrorLog, web.ErrorHandlerFunc(requestHandler))))
	})
	if err != nil {
		log.Println(err)
		return
	}
}
