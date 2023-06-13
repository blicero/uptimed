// /home/krylon/go/src/github.com/blicero/uptimed/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-13 17:56:18 krylon>

package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/database"
	"github.com/blicero/uptimed/dnssd"
	"github.com/blicero/uptimed/logdomain"
	"github.com/gorilla/mux"
	"github.com/wcharczuk/go-chart/v2"
)

const poolSize = 4

//go:embed assets
var assets embed.FS // nolint: unused

// Server wraps the http server and its associated state.
type Server struct {
	Address   string
	Port      int
	web       http.Server
	log       *log.Logger
	router    *mux.Router
	tmpl      *template.Template
	mimeTypes map[string]string
	pool      *database.Pool
	lock      sync.RWMutex
	period    int64
	hosts     map[string]time.Time
	mdns      *dnssd.Server
}

// Open creates a new Server.
func Open(addr string, port int) (*Server, error) {
	var (
		err error
		srv = &Server{
			Port:   port,
			period: 24,
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
			hosts: make(map[string]time.Time),
		}
		hostname string
	)

	if hostname, err = os.Hostname(); err != nil {
		return nil, err
	} else if srv.log, err = common.GetLogger(logdomain.Web); err != nil {
		return nil, err
	} else if srv.pool, err = database.NewPool(poolSize); err != nil {
		srv.log.Printf("[ERROR] Cannot create DB pool: %s\n",
			err.Error())
		return nil, err
	} else if srv.mdns, err = dnssd.CreateService(hostname, port); err != nil {
		srv.log.Printf("[ERROR] Cannot register service with mDNS: %s\n",
			err.Error())
		return nil, err
	}

	const tmplFolder = "assets/templates"
	var (
		templates []fs.DirEntry
		tmplRe    = regexp.MustCompile("[.]tmpl$")
		msg       string
	)

	if templates, err = assets.ReadDir(tmplFolder); err != nil {
		srv.log.Printf("[ERROR] Cannot read embedded templates: %s\n",
			err.Error())
		return nil, err
	}

	srv.tmpl = template.New("").Funcs(funcmap)
	for _, entry := range templates {
		var (
			content []byte
			path    = filepath.Join(tmplFolder, entry.Name())
		)

		if !tmplRe.MatchString(entry.Name()) {
			continue
		} else if content, err = assets.ReadFile(path); err != nil {
			msg = fmt.Sprintf("Cannot read embedded file %s: %s",
				path,
				err.Error())
			srv.log.Printf("[CRITICAL] %s\n", msg)
			return nil, errors.New(msg)
		} else if srv.tmpl, err = srv.tmpl.Parse(string(content)); err != nil {
			msg = fmt.Sprintf("Could not parse template %s: %s",
				entry.Name(),
				err.Error())
			srv.log.Println("[CRITICAL] " + msg)
			return nil, errors.New(msg)
		} else if common.Debug {
			srv.log.Printf("[TRACE] Template \"%s\" was parsed successfully.\n",
				entry.Name())
		}
	}

	// ...
	srv.router = mux.NewRouter()
	srv.web.Addr = addr
	srv.web.ErrorLog = srv.log
	srv.web.Handler = srv.router

	compress, _ := httpcompression.DefaultAdapter()

	srv.router.Use(compress)

	srv.router.HandleFunc("/{page:(?:main|start|index)?$}", srv.handleMain)
	srv.router.HandleFunc("/host/{hostname:(?:\\w+)$}", srv.handleHost)
	srv.router.HandleFunc("/prefs", srv.handlePrefs)

	srv.router.HandleFunc("/chart/{hostname:(?:\\w+)$}", srv.handleChart)

	srv.router.HandleFunc("/ws/report", srv.handleReport)
	srv.router.HandleFunc("/ajax/beacon", srv.handleBeacon)
	srv.router.HandleFunc("/ajax/set_period/{hours:(?:\\d+)$}", srv.handleSetPeriod)

	srv.router.HandleFunc("/favicon.ico", srv.handleFavIco)
	srv.router.HandleFunc("/static/{file}", srv.handleStaticFile)

	return srv, nil
} // func Open(addr string) (*Server, error)

// ListenAndServe enters the HTTP server's main loop, i.e.
// this method must be called for the Web frontend to handle
// requests.
func (srv *Server) ListenAndServe() {
	var err error

	defer srv.log.Println("[INFO] Web server is shutting down")
	defer srv.mdns.Shutdown()

	srv.log.Printf("[INFO] Web frontend is going online at %s\n", srv.Address)
	http.Handle("/", srv.router)

	if err = srv.web.ListenAndServe(); err != nil {
		if err.Error() != "http: Server closed" {
			srv.log.Printf("[ERROR] ListenAndServe returned an error: %s\n",
				err.Error())
		} else {
			srv.log.Println("[INFO] HTTP Server has shut down.")
		}
	}
} // func (srv *Server) ListenAndServe()

//////////////////////////////////////////////////////
// Frontend //////////////////////////////////////////
//////////////////////////////////////////////////////

func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	const tmplName = "main"

	var (
		err  error
		msg  string
		db   *database.Database
		tmpl *template.Template
		data = tmplDataMain{
			tmplDataBase: tmplDataBase{
				Title:     "Main",
				Debug:     common.Debug,
				URL:       r.URL.String(),
				Timestamp: time.Now(),
			},
		}
	)

	if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if data.Clients, err = db.HostGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to load Hosts from database: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Records, err = db.RecentGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to load recent data from database: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleHost(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)
	const tmplName = "single_host"

	var (
		err  error
		msg  string
		db   *database.Database
		tmpl *template.Template
		data = tmplDataHost{
			tmplDataBase: tmplDataBase{
				Debug:     common.Debug,
				URL:       r.URL.String(),
				Timestamp: time.Now(),
			},
		}
	)

	vars := mux.Vars(r)
	data.Hostname = vars["hostname"]
	data.Title = data.Hostname

	if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if data.Clients, err = db.HostGetAll(); err != nil {
		msg = fmt.Sprintf("Failed to load Hosts from database: %s",
			err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Records, err = db.RecordGetByHost(data.Hostname, time.Unix(0, 0)); err != nil {
		srv.log.Printf("[ERROR] Failed to query data for Host %s: %s\n",
			data.Hostname,
			err.Error())
	}

	srv.lock.RLock()
	data.Period = srv.period
	srv.lock.RUnlock()

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleHost(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleChart(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err           error
		msg, hostname string
		db            *database.Database
		records       []common.Record
		hours         int64
		yesterday     time.Time
		data          = tmplDataMain{
			tmplDataBase: tmplDataBase{
				Debug:     common.Debug,
				URL:       r.URL.String(),
				Timestamp: time.Now(),
			},
		}
	)

	if err = r.ParseForm(); err != nil {
		srv.log.Printf("[ERROR] Cannot parse form data: %s\n", err.Error())
	} else {
		var hstr = r.FormValue("hours")
		if hours, err = strconv.ParseInt(hstr, 10, 64); err != nil {
			hours = 0
			srv.log.Printf("[ERROR] Cannot parse duration %q: %s\n",
				hstr,
				err.Error())
		}
	}

	if hours == 0 {
		srv.lock.RLock()
		hours = srv.period
		srv.lock.RUnlock()
	}

	yesterday = time.Now().Add(time.Hour * time.Duration(hours) * -1)

	vars := mux.Vars(r)
	hostname = vars["hostname"]
	data.Title = fmt.Sprintf("Recent data for %s", hostname)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if records, err = db.RecordGetByHost(hostname, yesterday); err != nil {
		msg = fmt.Sprintf("Error getting data for Host %s: %s",
			hostname,
			err.Error())
		srv.log.Println("[ERROR] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	var (
		load       [3][]float64
		timestamps = make([]time.Time, len(records))
	)

	load[0] = make([]float64, len(records))
	load[1] = make([]float64, len(records))
	load[2] = make([]float64, len(records))

	for i, r := range records {
		timestamps[i] = r.Timestamp
		load[0][i] = r.Load[0]
		load[1][i] = r.Load[1]
		load[2][i] = r.Load[2]
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:           "Time",
			ValueFormatter: chart.TimeValueFormatterWithFormat(common.TimestampFormat),
		},
		YAxis: chart.YAxis{
			Name: "Load Average",
		},
		Title: hostname,
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: timestamps,
				YValues: load[0],
			},
			chart.TimeSeries{
				XValues: timestamps,
				YValues: load[1],
			},
			chart.TimeSeries{
				XValues: timestamps,
				YValues: load[2],
			},
		},
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(200)
	if err = graph.Render(chart.PNG, w); err != nil {
		srv.log.Printf("[ERROR] Cannot render chart for %s: %s\n",
			hostname,
			err.Error())
	}
} // func (srv *Server) handleChart(w http.ResponseWriter, r *http.Request)

func (srv *Server) handlePrefs(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	const tmplName = "prefs"

	var (
		err  error
		db   *database.Database
		msg  string
		tmpl *template.Template
		data = tmplDataPrefs{
			tmplDataBase: tmplDataBase{
				Title:     "Preferences",
				Timestamp: time.Now(),
				Debug:     common.Debug,
				URL:       r.URL.String(),
			},
		}
	)

	srv.lock.RLock()
	data.Period = srv.period
	srv.lock.RUnlock()

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	switch strings.ToLower(r.Method) {
	case "get":
		if data.Clients, err = db.HostGetAll(); err != nil {
			msg = fmt.Sprintf("Failed to load all hosts from database: %s",
				err.Error())
			srv.log.Printf("[ERROR] %s\n", msg)
			srv.sendErrorMessage(w, msg)
			return
		}

		tmpl = srv.tmpl.Lookup(tmplName)
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		if err = tmpl.Execute(w, &data); err != nil {
			msg = fmt.Sprintf("Error rendering template %q: %s",
				tmplName,
				err.Error())
			srv.sendErrorMessage(w, msg)
			return
		}
	case "post":
		if err = r.ParseForm(); err != nil {
			msg = fmt.Sprintf("Cannot parse form data: %s", err.Error())
			srv.log.Printf("[ERROR] %s\n", msg)
			srv.sendErrorMessage(w, msg)
			return
		}

		var periodStr = r.FormValue("period")
		var period int64

		if period, err = strconv.ParseInt(periodStr, 10, 64); err != nil {
			msg = fmt.Sprintf("Cannot parse Integer %q: %s",
				periodStr,
				err.Error())
			srv.log.Printf("[ERROR] %s\n",
				err.Error())
			srv.sendErrorMessage(w, msg)
			return
		}

		srv.lock.Lock()
		srv.period = period
		srv.lock.Unlock()
		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
} // func (srv *Server) handlePrefs(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////
// AJAX //////////////////////////////////////////////
//////////////////////////////////////////////////////

func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request) {
	var timestamp = time.Now().Format(common.TimestampFormat)
	const appName = common.AppName + " " + common.Version
	var jstr = fmt.Sprintf(`{ "Status": true, "Message": "%s", "Timestamp": "%s", "Hostname": "%s" }`,
		appName,
		timestamp,
		hostname())
	var response = []byte(jstr)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(200)
	w.Write(response) // nolint: errcheck,gosec
} // func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleSetPeriod(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err    error
		pstr   string
		hours  int64
		status bool
	)

	vars := mux.Vars(r)
	pstr = vars["hours"]

	if hours, err = strconv.ParseInt(pstr, 10, 64); err != nil {
		srv.log.Printf("[ERROR] Cannot parse new period %q: %s\n",
			pstr,
			err.Error())
	} else {
		srv.lock.Lock()
		srv.period = hours
		srv.lock.Unlock()
		status = true
	}

	var response = fmt.Sprintf(`{"Status": %t, "Message": "Yup"}`, status)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	if status {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
	}
	w.Write([]byte(response)) // nolint: errcheck
} // func (srv *Server) handleSetPeriod(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////
// Interaction with Clients //////////////////////////
//////////////////////////////////////////////////////

func (srv *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s from %s\n",
		r.URL.EscapedPath(),
		r.RemoteAddr)

	var (
		err  error
		db   *database.Database
		msg  string
		buf  bytes.Buffer
		rec  common.Record
		res  response
		body []byte
	)

	if _, err = io.Copy(&buf, r.Body); err != nil {
		res.Message = fmt.Sprintf("Failed to read HTTP request body: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n",
			res.Message)
		goto SEND_RESPONSE
	}

	body = buf.Bytes()

	if err = json.Unmarshal(body, &rec); err != nil {
		res.Message = fmt.Sprintf("Failed to decode JSON payload: %s\n%s",
			err.Error(),
			body)
		srv.log.Printf("[ERROR] %s\n",
			res.Message)
		goto SEND_RESPONSE
	}

	srv.lock.Lock()
	srv.hosts[rec.Hostname] = rec.Timestamp
	srv.lock.Unlock()

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if err = db.RecordAdd(&rec); err != nil {
		res.Message = fmt.Sprintf("Failed to add Record from %s to database: %s",
			rec.Hostname,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", res.Message)
		goto SEND_RESPONSE
	}

	res.Status = true

SEND_RESPONSE:
	res.Timestamp = time.Now()
	var rbuf []byte
	if rbuf, err = json.Marshal(&res); err != nil {
		srv.log.Printf("[ERROR] Error serializing response: %s\n",
			err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(200)
	if _, err = w.Write(rbuf); err != nil {
		msg = fmt.Sprintf("Failed to send result: %s",
			err.Error())
		srv.log.Println("[ERROR] " + msg)
	}
} // func (srv *Server) handleReport(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////
// General stuff /////////////////////////////////////
//////////////////////////////////////////////////////

func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s\n",
		request.URL.EscapedPath())

	const (
		filename = "assets/static/favicon.ico"
		mimeType = "image/vnd.microsoft.icon"
	)

	w.Header().Set("Content-Type", mimeType)

	if !common.Debug {
		w.Header().Set("Cache-Control", "max-age=7200")
	} else {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
	}

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(filename); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", filename)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close()
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request)

func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request) {
	// Since we control what static files the server has available, we
	// can easily map MIME type to slice. Soon.

	vars := mux.Vars(request)
	filename := vars["file"]
	path := filepath.Join("assets", "static", filename)

	var mimeType string

	srv.log.Printf("[TRACE] Delivering static file %s to client\n", filename)

	var match []string

	if match = common.SuffixPattern.FindStringSubmatch(filename); match == nil {
		mimeType = "text/plain"
	} else if mime, ok := srv.mimeTypes[match[1]]; ok {
		mimeType = mime
	} else {
		srv.log.Printf("[ERROR] Did not find MIME type for %s\n", filename)
	}

	w.Header().Set("Content-Type", mimeType)

	if common.Debug {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
	} else {
		w.Header().Set("Cache-Control", "max-age=7200")
	}

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(path); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", path)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close()
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request)

func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string) {
	html := `
<!DOCTYPE html>
<html>
  <head>
    <title>Internal Error</title>
  </head>
  <body>
    <h1>Internal Error</h1>
    <hr />
    We are sorry to inform you an internal application error has occured:<br />
    %s
    <p>
    Back to <a href="/index">Homepage</a>
    <hr />
    &copy; 2018 <a href="mailto:krylon@gmx.net">Benjamin Walkenhorst</a>
  </body>
</html>
`

	srv.log.Printf("[ERROR] %s\n", msg)

	output := fmt.Sprintf(html, msg)
	w.WriteHeader(500)
	_, _ = w.Write([]byte(output)) // nolint: gosec
} // func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string)
