package webui

import (
	"io/fs"
	"net/http"
)

func StaticHandler() http.Handler {
	sub, err := fs.Sub(FS, "static")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(sub))
}

func IndexHTML() []byte {
	data, err := FS.ReadFile("templates/index.html")
	if err != nil {
		return []byte(`<!DOCTYPE html><html><body><h1>RockPanel</h1><p>Frontend not loaded.</p></body></html>`)
	}
	return data
}