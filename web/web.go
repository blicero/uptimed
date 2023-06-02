// /home/krylon/go/src/github.com/blicero/uptimed/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-02 19:14:05 krylon>

package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/database"
	"github.com/blicero/uptimed/logdomain"
	"github.com/gorilla/mux"
)

const poolSize = 4

type Server struct {
	Address   string
	web       http.Server
	log       *log.Logger
	router    *mux.Router
	tmpl      *template.Template
	mimeTypes map[string]string
	pool      *database.Pool
}

func Open(addr string) (*Server, error) {
	var (
		err error
		srv = &Server{
			mimeTypes: map[string]string{
				".css":  "text/css",
				".map":  "application/json",
				".js":   "text/javascript",
				".png":  "image/png",
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".webp": "image/webp",
				".gif":  "image/gif",
				".json": "application/json",
				".html": "text/html",
			},
		}
	)

	if srv.log, err = common.GetLogger(logdomain.Web); err != nil {
		return nil, err
	} else if srv.pool, err = database.NewPool(poolSize); err != nil {
		srv.log.Printf("[ERROR] Cannot create DB pool: %s\n",
			err.Error())
		return nil, err
	}

	// ...
	srv.router = mux.NewRouter()
	srv.web.Addr = addr
	srv.web.ErrorLog = srv.log
	srv.web.Handler = srv.router

	//srv.router.HandleFunc("/ws/report",

	return srv, nil
} // func Open(addr string) (*Server, error)

func (srv *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s\n",
		r.URL.EscapedPath())

	var (
		err error
		db  *database.Database
		r   common.Record
	)

	if err = r.ParseForm(); err != nil {
		srv.log.Printf("[ERROR] Failed to parse form data: %s\n",
			err.Error())
		return
	}

} // func (srv *Server) handleReport(w http.ResponseWriter, r *http.Request)
