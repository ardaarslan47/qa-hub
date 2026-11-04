package web

import (
	"embed"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

var staticHandler = http.FileServerFS(staticFiles)
