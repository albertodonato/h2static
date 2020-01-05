package server

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/albertodonato/h2static/version"
)

// FileHandler is an http.Handler which serves static files under the specified
// filesystem.
type FileHandler struct {
	FileSystem

	template *DirectoryListingTemplate
}

// NewFileHandler returns a FileHandler for the specified filesystem.
func NewFileHandler(fileSystem FileSystem) *FileHandler {
	return &FileHandler{
		FileSystem: fileSystem,
		template:   NewDirectoryListingTemplate(),
	}
}

// ServeHTTP handles a request for the static file serve.
func (f FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	basePath := path.Clean(upath)
	fullPath, err := filepath.Abs(filepath.Join(f.FileSystem.Root, basePath))
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError)
		return
	}
	pathInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeHTTPError(w, http.StatusNotFound)
		} else {
			http.Error(w, "Failed getting path info", http.StatusInternalServerError)
		}
		return
	}
	if pathInfo.IsDir() {
		// if found, append the index suffix
		indexPath := f.findIndexSuffix(basePath)
		if indexPath == "" {
			// no index found, list directory content
			file, err := f.FileSystem.Open(basePath)
			if err != nil {
				writeHTTPError(w, http.StatusInternalServerError)
				return
			}
			f.writeDirListing(w, r, basePath, file)
			return
		}
		fullPath += indexPath
	}
	http.ServeFile(w, r, fullPath)
}

// Check if an index file exists for the directory, return its suffix.
func (f FileHandler) findIndexSuffix(dirPath string) string {
	for _, suffix := range []string{"/index.html", "/index.htm"} {
		indexPath := dirPath + suffix
		if _, err := f.FileSystem.OpenFile(indexPath); err == nil {
			return suffix
		}
	}
	return ""
}

func (f FileHandler) writeDirListing(w http.ResponseWriter, r *http.Request, path string, dir http.File) {
	var err error
	if strings.ToLower(r.Header.Get("Accept")) == "application/json" {
		err = f.template.RenderJSON(w, path, dir)
	} else {
		err = f.template.RenderHTML(w, path, dir)
	}
	if err != nil {
		http.Error(w, "Error listing directory", http.StatusInternalServerError)
		return
	}
}

// LoggingHandler wraps an http.Handler providing logging at startup.
type LoggingHandler struct {
	http.Handler
}

type loggingResponseWriter struct {
	http.ResponseWriter

	statusCode int
	length     int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// ServeHTTP logs server startup and serves via the configured handler.
func (h LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wr := newLoggingResponseWriter(w)
	h.Handler.ServeHTTP(wr, r)
	log.Printf(
		`%s %s %s %d %d %d "%s"`,
		r.Proto, r.Method, r.URL, r.ContentLength, wr.statusCode, wr.length,
		r.Header.Get("User-Agent"))
}

// BasicAuthHandler provides Basic Authorization.
type BasicAuthHandler struct {
	http.Handler

	// User/password pairs
	Credentials map[string]string
	// The authentication realm
	Realm string
}

// ServeHTTP logs server startup and serves via the configured handler.
func (h BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if user == "" {
		h.authRequiredResponse(w)
		return
	}
	hash, userFound := h.Credentials[user]
	if !ok || !userFound || h.hashPassword(pass) != hash {
		h.authRequiredResponse(w)
		return
	}
	h.Handler.ServeHTTP(w, r)
}

func (h BasicAuthHandler) authRequiredResponse(w http.ResponseWriter) {
	w.Header().Set(
		"WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, h.Realm))
	writeHTTPError(w, http.StatusUnauthorized)
}

func (h BasicAuthHandler) hashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// CommonHeadersHandler adds common headers to the response.
type CommonHeadersHandler struct {
	http.Handler
}

// ServeHTTP adds common headers to the response.
func (h CommonHeadersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Server", fmt.Sprintf("%s/%s", version.App.Name, version.App.Version))
	h.Handler.ServeHTTP(w, r)
}

func writeHTTPError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}
