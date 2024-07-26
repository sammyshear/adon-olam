package api

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sammyshear/adon-olam/views"
)

func MuxWithRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// content routes
	indexPage := views.Index()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", templ.Handler(indexPage))

	return mux
}
