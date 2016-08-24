//By Izan Beltrán Ferreiro - izanbf.es

package main

import (
	"strings"
)

var m map[string]string

func MIMEinit() {
	m = map[string]string {
		"html": "text/html",
		"php": "text/html",
		"htm": "text/html",
		"css": "text/css",
		"exe": "application/octet-stream",
		"png": "image/png",
		"jpg": "image/jpeg",
	}
}

func getMime (path *string) string {
	indexOfDot := strings.LastIndex(*path, ".")
	extension := (*path)[indexOfDot+1:]
	val, exists := m[extension]

	var mime string

	if !exists {
		return "text/plain"
	}

	mime = val

	return mime
}