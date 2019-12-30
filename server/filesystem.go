package server

import (
	"net/http"
	"os"
	"strings"
)

// HTMLPageResolveFileSystem provides a FileSystem which serves .html or .htm
// files for the corresponding path without suffix, if the original path is not
// found.
type HTMLPageResolveFileSystem struct {
	http.FileSystem
}

func (fs HTMLPageResolveFileSystem) Open(name string) (http.File, error) {
	file, err := fs.FileSystem.Open(name)
	if !os.IsNotExist(err) {
		return file, err
	}

	if !(strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm")) {
		newName := name
		for _, suffix := range []string{".html", ".htm"} {
			newName = name + suffix
			if file, err := fs.FileSystem.Open(newName); err == nil {
				if fileInfo, err := file.Stat(); err == nil && !fileInfo.IsDir() {
					return file, nil
				}
			}
		}
	}

	// return the result of the original call
	return file, err
}
