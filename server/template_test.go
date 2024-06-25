package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestDirectoryListingTemplate(t *testing.T) {
	suite.Run(t, new(DirectoryListingTemplateTestSuite))
}

type DirectoryListingTemplateTestSuite struct {
	testhelpers.TempDirTestSuite

	dir *server.File
}

func (s *DirectoryListingTemplateTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	fs := server.FileSystem{
		ResolveHTML:  true,
		HideDotFiles: true,
		Root:         s.TempDir,
	}
	dir, err := fs.Open("/")
	s.Nil(err)
	s.dir = dir
	s.WriteFile("foo", "foo content")
	s.WriteFile("bar", "bar content")
	s.Mkdir("baz")
}

// RenderHTML renders the HTML template.
func (s *DirectoryListingTemplateTestSuite) TestRenderHTML() {
	template := server.NewDirectoryListingTemplate(server.DirectoryListingTemplateConfig{})
	w := httptest.NewRecorder()
	template.RenderHTML(w, "/", s.dir, "", true)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a title="bar" href="bar" class="col col-name type-file" tabindex="1">bar</a>`)
	s.Contains(content, `<a title="baz/" href="baz/" class="col col-name type-dir" tabindex="2">baz/</a>`)
	s.Contains(content, `<a title="foo" href="foo" class="col col-name type-file" tabindex="3">foo</a>`)
}

// RenderHTML renders the HTML template with the correct path prefix.
func (s *DirectoryListingTemplateTestSuite) TestRenderHTMLWithPathPrefix() {
	template := server.NewDirectoryListingTemplate(server.DirectoryListingTemplateConfig{PathPrefix: "/prefix"})
	w := httptest.NewRecorder()
	template.RenderHTML(w, "/", s.dir, "", true)
	content := w.Body.String()
	s.Contains(content, `<link rel="shortcut icon" type="image/svg+xml" href="/prefix/.h2static-assets/logo.svg">`)
	s.Contains(content, `<link rel="stylesheet" type="text/css" href="/prefix/.h2static-assets/style.css">`)
}

// RenderHTML renders controls for descending sorting.
func (s *DirectoryListingTemplateTestSuite) TestRenderHTMLSortControlsDesc() {
	template := server.NewDirectoryListingTemplate(server.DirectoryListingTemplateConfig{})
	w := httptest.NewRecorder()
	template.RenderHTML(w, "/", s.dir, "", true)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a class="col col-name " href="?c=n&o=d">Name</a>`)
	s.Contains(content, `<a class="col col-size " href="?c=s&o=d">Size</a>`)
}

// RenderHTML renders controls for ascending sorting.
func (s *DirectoryListingTemplateTestSuite) TestRenderHTMLSortControlsAsc() {
	template := server.NewDirectoryListingTemplate(server.DirectoryListingTemplateConfig{})
	w := httptest.NewRecorder()
	template.RenderHTML(w, "/", s.dir, "", false)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("text/html; charset=utf-8", response.Header.Get("Content-Type"))
	content := w.Body.String()
	s.Contains(content, `<a class="col col-name " href="?c=n&o=a">Name</a>`)
	s.Contains(content, `<a class="col col-size " href="?c=s&o=a">Size</a>`)
}

// RenderJSON renders JSON listing.
func (s *DirectoryListingTemplateTestSuite) TestRenderJSON() {
	template := server.NewDirectoryListingTemplate(server.DirectoryListingTemplateConfig{})
	w := httptest.NewRecorder()
	template.RenderJSON(w, "/", s.dir, "", true)
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

func TestGetHumanByteSize(t *testing.T) {
	suite.Run(t, new(GetHumanByteSizeTestSuite))
}

type GetHumanByteSizeTestSuite struct {
	suite.Suite
}

// The value is converted with bytes.
func (s *GetHumanByteSizeTestSuite) TestBytes() {
	info := server.GetHumanByteSize(10)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("B", info.Suffix)
}

// The value is converted with Kilobytes.
func (s *GetHumanByteSizeTestSuite) TestKiloBytes() {
	info := server.GetHumanByteSize(10 * 1024)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("KB", info.Suffix)
}

// The value is converted with Megabytes.
func (s *GetHumanByteSizeTestSuite) TestMegaBytes() {
	info := server.GetHumanByteSize(10 * 1024 * 1024)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("MB", info.Suffix)
}

// The value is converted with Gigabytes.
func (s *GetHumanByteSizeTestSuite) TestGigaBytes() {
	info := server.GetHumanByteSize(10 * 1024 * 1024 * 1024)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("GB", info.Suffix)
}

// The value is converted with Terabytes.
func (s *GetHumanByteSizeTestSuite) TestTeraBytes() {
	info := server.GetHumanByteSize(10 * 1024 * 1024 * 1024 * 1024)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("TB", info.Suffix)
}

// The value is converted with Petabytes.
func (s *GetHumanByteSizeTestSuite) TestPetaBytes() {
	info := server.GetHumanByteSize(10 * 1024 * 1024 * 1024 * 1024 * 1024)
	s.Equal(server.FileSize(10.0), info.Value)
	s.Equal("PB", info.Suffix)
}

// Decimal part is include
func (s *GetHumanByteSizeTestSuite) TestWithDecimal() {
	info := server.GetHumanByteSize(int64(1.5 * 1024 * 1024 * 1024 * 1024 * 1024))
	s.Equal(server.FileSize(1.5), info.Value)
	s.Equal("PB", info.Suffix)
}
