package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
)

func TestDirectoryListingTemplate(t *testing.T) {
	suite.Run(t, new(DirectoryListingTemplateTestSuite))
}

type DirectoryListingTemplateTestSuite struct {
	TempDirTestSuite

	dir http.File
}

func (s *DirectoryListingTemplateTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	dir, err := http.Dir(s.TempDir).Open("/")
	s.Nil(err)
	s.dir = dir
	s.WriteFile("foo", "foo content")
	s.WriteFile("bar", "bar content")
	s.Mkdir("baz")
}

// RenderHTML renders the HTML template.
func (s *DirectoryListingTemplateTestSuite) TestRenderHTML() {
	template := server.NewDirectoryListingTemplate()
	w := httptest.NewRecorder()
	template.RenderHTML(w, "/", s.dir)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a href="foo" class="button link type-file">foo</a>`)
	s.Contains(content, `<a href="bar" class="button link type-file">bar</a>`)
	s.Contains(content, `<a href="baz/" class="button link type-dir">baz/</a>`)
}

// RenderJSON renders JSON listing.
func (s *DirectoryListingTemplateTestSuite) TestRenderJSON() {
	template := server.NewDirectoryListingTemplate()
	w := httptest.NewRecorder()
	template.RenderJSON(w, "/", s.dir)
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
					Name:  "baz/",
					IsDir: true,
					Size:  0,
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
