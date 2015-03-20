// main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/JamesDunne/go-util/base"
	"github.com/JamesDunne/go-util/web"
	"net"
	"net/http"
)

import _ "github.com/lib/pq"

var db *sql.DB

const (
	dbConnectionString = "host=localhost sslmode=disable dbname=band user=band password=band"
)

func daemon() error {
	select {}
}

func main() {
	fl_listen_uri := flag.String("l", "tcp://0.0.0.0:8080", "listen URI (schemes available are tcp, unix)")
	flag.Parse()

	// Parse all the URIs:
	listen_addr, err := base.ParseListenable(*fl_listen_uri)
	base.PanicIf(err)

	// Open database:
	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Startup")
	base.ServeMain(listen_addr, func(l net.Listener) error {
		return http.Serve(l, web.ReportErrors(web.Log(web.DefaultErrorLog, web.ErrorHandlerFunc(requestHandler))))
	})
	fmt.Println("Shutdown")
}
