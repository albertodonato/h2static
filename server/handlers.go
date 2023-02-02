package server

import (
	"crypto/sha512"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

// FileHandler is an http.Handler which serves static files under the specified
// filesystem.
type FileHandler struct {
	FileSystem     FileSystem
	DirectoryIndex bool
	pathPrefix     string
	template       *DirectoryListingTemplate
}

// NewFileHandler returns a FileHandler for the specified filesystem.
func NewFileHandler(fileSystem FileSystem, directoryIndex bool, pathPrefix string) *FileHandler {
	return &FileHandler{
		FileSystem:     fileSystem,
		DirectoryIndex: directoryIndex,
		pathPrefix:     pathPrefix,
		template:       NewDirectoryListingTemplate(DirectoryListingTemplateConfig{PathPrefix: pathPrefix}),
	}
}

// ServeHTTP handles a request for the static file serve.
func (f FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if method := strings.ToUpper(r.Method); method != http.MethodGet && method != http.MethodHead {
		writeHTTPError(w, http.StatusMethodNotAllowed)
		return
	}

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
		} else if os.IsPermission(err) {
			writeHTTPError(w, http.StatusForbidden)
		} else {
			writeServerError(w, err)
		}
		return
	}
	fullPath := file.AbsPath()
	if file.Info.IsDir() {
		if !strings.HasSuffix(urlPath, "/") {
			// always redirect to URL with trailing slash for directories
			localRedirect(w, r, f.pathPrefix+urlPath+"/")
			return
		}
		// if found, append the index suffix
		indexPath := f.findIndexSuffix(basePath)
		if indexPath == "" {
			if !f.DirectoryIndex {
				// directory listing disallowed
				writeHTTPError(w, http.StatusForbidden)
				return
			}

			// list directory content
			file, err := f.FileSystem.Open(basePath)
			if err != nil {
				writeServerError(w, err)
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
		writeServerError(w, err)
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
	remoteAddr := h.getRemoteAddr(r)
	log.Printf(
		`%s %s %s %d %d %d %s "%s"`,
		r.Proto, r.Method, r.URL, r.ContentLength, wr.statusCode, wr.length, remoteAddr,
		r.Header.Get("User-Agent"))
}

func (h LoggingHandler) getRemoteAddr(r *http.Request) string {
	addr := r.RemoteAddr
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		addresses := strings.Split(xff, ",")
		xffAddr := strings.Trim(addresses[len(addresses)-1], " ")
		addr = fmt.Sprintf("%s [%s]", addr, xffAddr)
	}
	return addr
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

// AddHeadersHandler wraps an http.Handler adding headers.
func AddHeadersHandler(headers map[string]string, h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			for key, value := range headers {
				header.Set(key, value)
			}
			h.ServeHTTP(w, r)
		})
}

//go:embed assets
var assetsFileSystem embed.FS

// AssetsHandler serves static assets for the server.
func AssetsHandler() http.Handler {
	fileSystem, _ := fs.Sub(assetsFileSystem, "assets")
	return http.FileServer(http.FS(fileSystem))
}

func writeHTTPError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

func writeServerError(w http.ResponseWriter, err error) {
	log.Printf("Error: %v", err)
	writeHTTPError(w, http.StatusInternalServerError)
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
