package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
	"strconv"

	"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
	_ "github.com/lib/pq"
)

const (
	defaultDbConnectionString = "host=localhost sslmode=disable dbname=band user=band password=band"
)

var (
	html_folder        = ""
	staticHref         = ""
	verbose            = false
	debug              = false
	failsafe           = false
	dbConnectionString string
	db                 *sql.DB
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

var uiTmpl *template.Template

func main() {
	// Define our commandline flags:
	flag.StringVar(&html_folder, "html", "./html", "Directory of HTML template files")
	flag.StringVar(&staticHref, "static", "static", "HREF prefix to static files")
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available are tcp, unix)")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&debug, "d", false, "debug logging")
	flag.BoolVar(&failsafe, "f", false, "failsafe mode - SQL errors return empty resultsets")
	// TODO: configurable DB connection string
	flag.StringVar(&dbConnectionString, "db", defaultDbConnectionString, "DB connection string")
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
