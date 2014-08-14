package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
)

import (
	"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
)

var (
	base_folder = ""
)

func html_path() string { return base_folder + "/html" }

var uiTmpl *template.Template

func main() {
	// Define our commandline flags:
	fs := flag.String("fs", ".", "Root directory of served files and templates")
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available are tcp, unix)")
	flag.Parse()

	// Parse all the URIs:
	listen_addr, err := base.ParseListenable(*fl_listen_uri)
	base.PanicIf(err)

	// Make directories we need:
	base_folder = base.CanonicalPath(path.Clean(*fs))

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
