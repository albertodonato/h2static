package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/version"
)

func TestFileHandler(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}

type FileHandlerTestSuite struct {
	TempDirTestSuite

	fileSystem server.FileSystem
	handler    *server.FileHandler
}

func (s *FileHandlerTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	s.fileSystem = server.NewFileSystem(s.TempDir, true, true)
	s.handler = server.NewFileHandler(s.fileSystem)
	s.WriteFile("foo", "foo content")
	s.WriteFile("bar", "bar content")
	s.Mkdir("baz")
}

func (s *FileHandlerTestSuite) TestListingHTML() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a href="foo" class="button link type-file">foo</a>`)
	s.Contains(content, `<a href="bar" class="button link type-file">bar</a>`)
	s.Contains(content, `<a href="baz/" class="button link type-dir">baz/</a>`)
	// The root directory doesn't contain a link up
	s.NotContains(content, `<a href=".." class="button link type-dir-up">..</a>`)
}

func (s *FileHandlerTestSuite) TestListingHTMLSubdir() {
	r := httptest.NewRequest("GET", "/baz", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a href=".." class="button link type-dir-up">..</a>`)
}

func (s *FileHandlerTestSuite) TestServeFile() {
	r := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/plain; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Equal("foo content", content)
}

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

func (s *FileHandlerTestSuite) TestListingJSON() {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("application/json", response.Header.Get("Content-Type"))
	decoder := json.NewDecoder(w.Body)
	content := server.DirInfo{}
	decoder.Decode(&content)
	s.Equal(
		server.DirInfo{
			Name:   "/",
			IsRoot: true,
			Entries: []server.DirEntryInfo{
				{
					Name:  "bar",
					IsDir: false,
					Size:  11,
				},
				{
					Name:  "baz",
					IsDir: true,
					Size:  s.Stat("baz").Size(),
				},
				{
					Name:  "foo",
					IsDir: false,
					Size:  11,
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

func (s *BasicAuthHandlerTestSuite) TestValidCredentials() {
	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("foo", "bar")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusNotFound, response.StatusCode)
}

func TestCommonHeadersHandler(t *testing.T) {
	suite.Run(t, new(CommonHeadersHandlerTestSuite))
}

type CommonHeadersHandlerTestSuite struct {
	suite.Suite
}

func (s *CommonHeadersHandlerTestSuite) TestAddHeaders() {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := server.CommonHeadersHandler{Handler: http.NotFoundHandler()}
	handler.ServeHTTP(w, r)
	response := w.Result()
	server := fmt.Sprintf("%s/%s", version.App.Name, version.App.Version)
	s.Equal(http.StatusNotFound, response.StatusCode)
	s.Equal(server, response.Header.Get("Server"))
}
