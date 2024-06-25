package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestFileHandler(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}

type FileHandlerTestSuite struct {
	testhelpers.TempDirTestSuite

	fileSystem server.FileSystem
	handler    *server.FileHandler
}

func (s *FileHandlerTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	s.fileSystem = server.FileSystem{
		Root:         s.TempDir,
		ResolveHTML:  true,
		HideDotFiles: true,
	}
	s.handler = server.NewFileHandler(s.fileSystem, true, "")
	s.WriteFile("foo", "foofoofoo")
	s.WriteFile("bar", "barbar")
	s.Mkdir("baz")
}

// HEAD requests are replied
func (s *FileHandlerTestSuite) TestHEADRequest() {
	r := httptest.NewRequest("HEAD", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
}

// Methods others than GET and HEAD are not allowed
func (s *FileHandlerTestSuite) TestMethodNotAllowed() {
	r := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusMethodNotAllowed, response.StatusCode)
}

// HTML listing contains links to entries.
func (s *FileHandlerTestSuite) TestListingHTML() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a title="bar" href="bar" class="col col-name type-file" tabindex="1">bar</a>`)
	s.Contains(content, `<a title="baz/" href="baz/" class="col col-name type-dir" tabindex="2">baz/</a>`)
	s.Contains(content, `<a title="foo" href="foo" class="col col-name type-file" tabindex="3">foo</a>`)
	// The root directory doesn't contain a link up
	s.NotContains(content, `<a href=".." class="col col-name type-dir-up">..</a>`)
}

// HTML listing can be disallowed
func (s *FileHandlerTestSuite) TestListingHTMLDisallowed() {
	handler := server.NewFileHandler(s.fileSystem, false, "")
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusForbidden, response.StatusCode)
	content := w.Body.String()
	s.Contains(content, "403 Forbidden")
}

// HTML listing for a subdirectory has a link to the parent.
func (s *FileHandlerTestSuite) TestListingHTMLSubdir() {
	r := httptest.NewRequest("GET", "/baz/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a title="Up one directory" href=".." class="col col-name type-dir-up">..</a>`)
}

// File content is served.
func (s *FileHandlerTestSuite) TestServeFile() {
	r := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/plain; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Equal("foofoofoo", content)
}

// URLs for directories without trailing shash are redirected to the URL with
// slash.
func (s *FileHandlerTestSuite) TestDirectoryRedirectWithTrailingSlash() {
	r := httptest.NewRequest("GET", "/baz", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusMovedPermanently, response.StatusCode)
	s.Equal("/baz/", response.Header.Get("Location"))
}

// URLs for directories without trailing shash are redirected to the URL with
// slash, including the path prefix.
func (s *FileHandlerTestSuite) TestDirectoryRedirectWithTrailingSlashPathPrefix() {
	handler := server.NewFileHandler(s.fileSystem, false, "/prefix")
	r := httptest.NewRequest("GET", "/baz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusMovedPermanently, response.StatusCode)
	s.Equal("/prefix/baz/", response.Header.Get("Location"))
}

// If a directory has an index.html file, it's served instead of listing.
func (s *FileHandlerTestSuite) TestServeDirectoryIndexHTML() {
	s.WriteFile("index.html", "some content")
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Equal("some content", content)
}

// If a directory has an index.htm file, it's served instead of listing.
func (s *FileHandlerTestSuite) TestServeDirectoryIndexHTM() {
	s.WriteFile("index.htm", "some content")
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Equal("some content", content)
}

// The index.html file is preferred to index.htm
func (s *FileHandlerTestSuite) TestServeDirectoryPreferIndexHTML() {
	s.WriteFile("index.html", "some content")
	s.WriteFile("index.htm", "other content")
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Equal("some content", content)
}

// JSON listing is returned if the Accept header is set.
func (s *FileHandlerTestSuite) TestListingJSON() {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("application/json", response.Header.Get("Content-Type"))
	decoder := json.NewDecoder(w.Body)
	var content server.DirInfo
	decoder.Decode(&content)
}

// JSON listing can be sorted in descending order.
func (s *FileHandlerTestSuite) TestListingJSONSortDesc() {
	s.RemoveAll("/baz")
	r := httptest.NewRequest("GET", "/?o=d", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("application/json", response.Header.Get("Content-Type"))
	decoder := json.NewDecoder(w.Body)
	var content server.DirInfo
	decoder.Decode(&content)
	s.Equal(
		server.DirInfo{
			Name:   "/",
			IsRoot: true,
			Entries: []server.DirEntryInfo{
				{
					Name:  "foo",
					IsDir: false,
					Size:  9,
				},
				{
					Name:  "bar",
					IsDir: false,
					Size:  6,
				},
			},
		},
		content,
	)
}

// JSON listing can be sorted by size.
func (s *FileHandlerTestSuite) TestListingJSONSortSize() {
	s.RemoveAll("/baz")
	r := httptest.NewRequest("GET", "/?c=s", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("application/json", response.Header.Get("Content-Type"))
	decoder := json.NewDecoder(w.Body)
	var content server.DirInfo
	decoder.Decode(&content)
	s.Equal(
		server.DirInfo{
			Name:   "/",
			IsRoot: true,
			Entries: []server.DirEntryInfo{
				{
					Name:  "bar",
					IsDir: false,
					Size:  6,
				},
				{
					Name:  "foo",
					IsDir: false,
					Size:  9,
				},
			},
		},
		content,
	)
}

func TestBasicAuthHandler(t *testing.T) {
	suite.Run(t, new(BasicAuthHandlerTestSuite))
}

type BasicAuthHandlerTestSuite struct {
	suite.Suite

	handler server.BasicAuthHandler
}

func (s *BasicAuthHandlerTestSuite) SetupTest() {
	s.handler = server.BasicAuthHandler{
		Handler: http.NotFoundHandler(), // a valid request returns 404
		Credentials: map[string]string{
			// password is "bar"
			"foo": "d82c4eb5261cb9c8aa9855edd67d1bd10482f41529858d925094d173fa662aa91ff39bc5b188615273484021dfb16fd8284cf684ccf0fc795be3aa2fc1e6c181"},
		Realm: "tests",
	}
}

// A 401 response is returned if no credentials are provided.
func (s *BasicAuthHandlerTestSuite) TestNoCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusUnauthorized, response.StatusCode)
	s.Equal("401 Unauthorized", response.Status)
	s.Equal(
		`Basic realm="tests", charset="UTF-8"`,
		response.Header.Get("WWW-Authenticate"))
}

// A 401 response is returned if credentials are invalid.
func (s *BasicAuthHandlerTestSuite) TestInvalidCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("foo", "wrong")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusUnauthorized, response.StatusCode)
	s.Equal("401 Unauthorized", response.Status)
	s.Equal(
		`Basic realm="tests", charset="UTF-8"`,
		response.Header.Get("WWW-Authenticate"))
}

// A response is returned if credentials match.
func (s *BasicAuthHandlerTestSuite) TestValidCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("foo", "bar")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
}

func TestAddHeadersHandler(t *testing.T) {
	suite.Run(t, new(AddHeadersHandlerTestSuite))
}

type AddHeadersHandlerTestSuite struct {
	suite.Suite
}

// Specified headers are added to the response.
func (s *AddHeadersHandlerTestSuite) TestAddHeaders() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := server.AddHeadersHandler(
		map[string]string{"X-Foo": "foo", "X-Bar": "bar"},
		http.NotFoundHandler())
	handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
	s.Equal("foo", response.Header.Get("X-Foo"))
	s.Equal("bar", response.Header.Get("X-Bar"))
}

func TestAssetsHandler(t *testing.T) {
	suite.Run(t, new(AssetsHandlerTestSuite))
}

type AssetsHandlerTestSuite struct {
	suite.Suite
}

// Static assets are served with the right content-type.
func (s *AssetsHandlerTestSuite) TestServeAssets() {
	handler := server.AssetsHandler()
	filesMap := map[string]string{
		"logo.svg":  "image/svg+xml",
		"style.css": "text/css; charset=utf-8",
	}
	for filePath, contentType := range filesMap {
		r := httptest.NewRequest("GET", "/"+filePath, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()
		s.Equal(http.StatusOK, response.StatusCode)
		s.Equal(contentType, response.Header.Get("Content-Type"))
	}
}

func TestLoggingHandler(t *testing.T) {
	suite.Run(t, new(LoggingHandlerTestSuite))
}

type LoggingHandlerTestSuite struct {
	testhelpers.LogCaptureSuite

	handler server.LoggingHandler
}

func (s *LoggingHandlerTestSuite) SetupTest() {
	s.LogCaptureSuite.SetupTest()
	s.handler = server.LoggingHandler{Handler: http.NotFoundHandler()}
}

// Requests are logged.
func (s *LoggingHandlerTestSuite) TestLogRequest() {
	r := httptest.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	s.Contains(s.Logs.String(), "HTTP/1.1 GET /path 0 404 19 "+r.RemoteAddr)
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
}

// Requests are logged with the original request IP.
func (s *LoggingHandlerTestSuite) TestLogRequestWithForward() {
	r := httptest.NewRequest("GET", "/path", nil)
	r.Header.Add("X-Forwarded-For", "1.1.1.1, 8.8.8.8, 1.2.3.4")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	s.Contains(
		s.Logs.String(),
		fmt.Sprintf("HTTP/1.1 GET /path 0 404 19 %s [1.2.3.4]", r.RemoteAddr))
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
}
