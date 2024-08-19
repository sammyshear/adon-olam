package api

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/sammyshear/adon-olam/views"
)

func MuxWithRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	ch := make(chan channel)
	wg := &sync.WaitGroup{}

	go uploadMidiProcessor(ch, wg)

	// content routes
	indexPage := views.Index()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", templ.Handler(indexPage))

	// api routes
	mux.HandleFunc("POST /api/upload", UploadMidiHandler(ch))
	mux.HandleFunc("GET /api/status/{requestID}", JobStatusHandler)
	mux.HandleFunc("GET /api/status/{requestID}/tick", JobStatusTicker)

	return mux
}
