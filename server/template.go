package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/albertodonato/h2static/version"
)

// template for the directory listing page
var dirListingTemplateText = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .App.Name }} - Directory listing for {{ .Dir.Name }}</title>
    <style type="text/css">
      body {
        width: 90%;
        margin: 0 auto;
        font-family: sans;
        font-size: 34px;
        color: black;
      }
      h1 {
        margin: 1em 0;
        font-size: 130%;
      }
      a, a:visited {
        color: black;
        text-decoration: none;
      }
      a:active, a:hover {
        color: #007bff;
        text-decoration: none;
      }
      .listing {
        width: 100%;
      }
      .entry {
        padding: 0.5rem 0;
        display: flex;
        justify-content: space-between;
      }
      .button {
        display: inline-block;
        margin: 0 0.2rem;
        padding: 1rem 0.5rem;
        font-family: monospace;
        border-width: 1px;
        border-style: solid;
        border-radius: 0.25rem;
        white-space: nowrap;
      }
      a.type-dir-up {
        flex-grow: 0;
        width: auto;
        background: #6c757d linear-gradient(to bottom, #828a91 0, #6c757d 100%);
        border-color: #6c757d;
        color: white;
      }
      a.type-dir {
        background: #337ab7 linear-gradient(to bottom, #337ab7 0, #2e6da4 100%);
        border-color: #337ab7;
        color: white;
      }
      a.type-file {
        background: #dddddd linear-gradient(to bottom, #f5f5f5 0, #e8e8e8 100%);
        border-color: #dddddd;
        color: #515151;
      }
      .path {
        font-family: monospace;
      }
      .link {
        flex-grow: 1;
      }
      .size {
        border-color: #777777;
        background-color: white;
        color: #777777;
        text-align: right;
        width: 11rem;
      }
      .size-suffix {
        display: inline-block;
        width: 1.5em;
        margin-left: 0.25em;
        font-size: 80%;
        text-align: left;
      }
      .powered-by {
        margin: 3em 0;
        text-align: center;
        font-size: 80%;
      }
      .powered-by a {
        font-family: monospace;
        font-size: 120%;
        margin-left: 0.5em;
      }
      a.powered-by:hover {
        text-decoration: underline;
      }

      @media (min-width: 992px) {
        body {
          width: 60%;
          font-size: 16px;
        }
        .button {
          padding: 0.5rem;
        }
        .size {
          width: 5rem;
        }
      }
    </style>
  </head>
  <body>
    <header>
      <h1>Directory listing for <span class="path">{{ .Dir.Name }}</span></h1>
    </header>
    <main>
      <section class="listing">
        {{- if .Dir.IsRoot -}}
        {{- else -}}
        <div class="entry">
          <a href=".." class="button link type-dir-up">..</a>
        </div>
        {{- end -}}
        {{- range .Dir.Entries -}}
        <div class="entry">
          {{ if .IsDir -}}
            <a href="{{ .Name }}/" class="button link type-dir">{{ .Name }}/</a>
          {{- else -}}
            <a href="{{ .Name }}" class="button link type-file">{{ .Name }}</a>
          {{- end }}
          <span class="button size">
            {{ .HumanSize.Value -}}
            <span class="size-suffix">{{ .HumanSize.Suffix }}</span>
          </span>
        </div>
        {{ end -}}
      </section>
    </main>
    <footer>
      <div class="powered-by">
        Powered by
        <a href="https://github.com/albertodonato/h2static">
          {{ .App.Name }} {{ .App.Version }}
        </a>
      </div>
    </footer>
  </body>
</html>
`

// DirInfo holds details about the directory being listed.
type DirInfo struct {
	Name    string
	IsRoot  bool
	Entries []DirEntryInfo
}

// DirEntryInfo holds details for a directory entry.
type DirEntryInfo struct {
	Name      string
	IsDir     bool
	Size      int64
	HumanSize humanSizeInfo `json:"-"`
}

// FileSize represent a file size as a float number.
type FileSize float64

// String returns a string representation of the FileSize.
func (f FileSize) String() string {
	return strings.TrimSuffix(fmt.Sprintf("%.1f", f), ".0")
}

type humanSizeInfo struct {
	Value  FileSize
	Suffix string
}

type templateContext struct {
	App version.Version
	Dir DirInfo
}

// DirectoryListingTemplate is a template rendered for a directory.
type DirectoryListingTemplate struct {
	template *template.Template
}

// NewDirectoryListingTemplate returns a DirectoryListingTemplate for the specified directory.
func NewDirectoryListingTemplate() *DirectoryListingTemplate {
	return &DirectoryListingTemplate{
		template: template.Must(
			template.New("DirListing").Parse(dirListingTemplateText)),
	}
}

// RenderHTML renders the HTML template for a directory.
func (t *DirectoryListingTemplate) RenderHTML(w http.ResponseWriter, path string, dir *File) error {
	context, err := t.getTemplateContext(path, dir)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.template.Execute(w, context)
}

// RenderJSON returns JSON listing for a directory.
func (t *DirectoryListingTemplate) RenderJSON(w http.ResponseWriter, path string, dir *File) error {
	context, err := t.getTemplateContext(path, dir)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(context.Dir)
}

// return directory info for the template
func (t *DirectoryListingTemplate) getTemplateContext(path string, dir *File) (*templateContext, error) {
	pathInfos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	toLower := func(info os.FileInfo) string {
		return strings.ToLower(info.Name())
	}

	sort.Slice(
		pathInfos,
		func(i, j int) bool {
			return toLower(pathInfos[i]) < toLower(pathInfos[j])
		})

	entries := []DirEntryInfo{}
	for _, p := range pathInfos {
		name := string(template.URL(p.Name()))
		size := p.Size()
		entries = append(entries, DirEntryInfo{
			Name:      name,
			IsDir:     p.IsDir(),
			Size:      size,
			HumanSize: getHumanByteSize(size),
		})
	}
	return &templateContext{
		App: version.App,
		Dir: DirInfo{
			Name:    path,
			IsRoot:  path == "/",
			Entries: entries,
		},
	}, nil
}

func getHumanByteSize(size int64) humanSizeInfo {
	value := FileSize(size)
	suffix := ""
	for _, s := range []string{"K", "M", "G", "T", "P", "E"} {
		if value < 1024 {
			break
		}
		value, suffix = value/1024, s
	}
	return humanSizeInfo{
		Value:  value,
		Suffix: suffix + "B",
	}
}
