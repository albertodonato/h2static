package server

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/albertodonato/h2static/version"
)

// FileHandler is an http.Handler which serves static files under the specified
// filesystem.
type FileHandler struct {
	FileSystem FileSystem
	template   *DirectoryListingTemplate
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
	urlPath := r.URL.Path
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
		r.URL.Path = urlPath
	}
	basePath := path.Clean(urlPath)
	file, err := f.FileSystem.Open(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			writeHTTPError(w, http.StatusNotFound)
		} else {
			writeHTTPError(w, http.StatusInternalServerError)
		}
		return
	}
	pathInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed getting path info", http.StatusInternalServerError)
		return
	}
	fullPath := file.AbsPath
	if pathInfo.IsDir() {
		if !strings.HasSuffix(urlPath, "/") {
			// always redirect to URL with trailing slash for directories
			localRedirect(w, r, urlPath+"/")
			return
		}
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

const sortColumns = "ns" // name, size

func (f FileHandler) writeDirListing(w http.ResponseWriter, r *http.Request, path string, dir *File) {
	q := r.URL.Query()
	sortColumn := q.Get("c")
	if !strings.Contains(sortColumns, sortColumn) || sortColumn == "" {
		sortColumn = "n"
	}
	sortAsc := q.Get("o") != "d"
	var err error
	if strings.ToLower(r.Header.Get("Accept")) == "application/json" {
		err = f.template.RenderJSON(w, path, dir, sortColumn, sortAsc)
	} else {
		err = f.template.RenderHTML(w, path, dir, sortColumn, sortAsc)
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

// AssetsHandler serves static assets.
type AssetsHandler struct {
	http.Handler

	Assets StaticAssets
}

// ServeHTTP return content for static assets.
func (h AssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	asset, ok := h.Assets[strings.TrimPrefix(r.URL.Path, "/")]
	if ok {
		header := w.Header()
		header.Set("Content-Type", asset.ContentType)
		header.Set("Cache-Control", "public, max-age=86400")
		w.Write(asset.Content)
		return
	}
	writeHTTPError(w, http.StatusNotFound)
}

func writeHTTPError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

// localRedirect gives a Moved Permanently response.  It does not convert
// relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
