package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func RegisterHandlers(mux *http.ServeMux) {
	sub, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/dashboard/", http.StripPrefix("/dashboard/", http.FileServer(http.FS(sub))))
}
