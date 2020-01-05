package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/albertodonato/h2static/version"
)

// AssetsPrefix defines the URL prefix for static assets.
const AssetsPrefix = "/.h2static-assets/"

// template for the directory listing page
var dirListingTemplateText = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .App.Name }} - Directory listing for {{ .Dir.Name }}</title>
    <link rel="stylesheet" type="text/css" href="{{ .AssetsPrefix }}style.css">
  </head>
  <body>
    <header>
      <h1>
        <a class="logo" alt="{{ .App.Name }}" href="/">
          <img src="{{ .AssetsPrefix }}logo.svg">
        </a>
        <span class="title">Directory listing for <span class="path">{{ .Dir.Name }}</span></span>
      </h1>
    </header>
    <main>
      <section class="listing">
        <div class="entry">
          {{- if .Sort.Asc -}}
          <a class="button link sort" href="?c=n&o=d">Name{{ if eq .Sort.Column "n" }} &#x25B2;{{ end }}</a>
          <a class="button size sort" href="?c=s&o=d">{{ if eq .Sort.Column "s" }}&#x25B2; {{ end }}Size</a>
          {{- else -}}
          <a class="button link sort" href="?c=n&o=a">Name{{ if eq .Sort.Column "n" }} &#x25BC;{{ end }}</a>
          <a class="button size sort" href="?c=s&o=a">{{ if eq .Sort.Column "s" }}&#x25BC; {{ end }}Size</a>
          {{- end -}}
        </div>
        {{- if not .Dir.IsRoot -}}
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

type sortInfo struct {
	Column string
	Asc    bool
}

type templateContext struct {
	App          version.Version
	AssetsPrefix string
	Sort         sortInfo
	Dir          DirInfo
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
func (t *DirectoryListingTemplate) RenderHTML(w http.ResponseWriter, path string, dir *File, sortColumn string, sortAsc bool) error {
	context, err := t.getTemplateContext(path, dir, sortColumn, sortAsc)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.template.Execute(w, context)
}

// RenderJSON returns JSON listing for a directory.
func (t *DirectoryListingTemplate) RenderJSON(w http.ResponseWriter, path string, dir *File, sortColumn string, sortAsc bool) error {
	context, err := t.getTemplateContext(path, dir, sortColumn, sortAsc)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(context.Dir)
}

// return directory info for the template
func (t *DirectoryListingTemplate) getTemplateContext(path string, dir *File, sortColumn string, sortAsc bool) (*templateContext, error) {
	pathInfos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var sortFunc func(int, int) bool
	if sortColumn == "s" { // sort by size
		sortFunc = func(i, j int) bool {
			return pathInfos[i].Size() < pathInfos[j].Size()
		}
	} else { // sort by name
		sortFunc = func(i, j int) bool {
			return strings.ToLower(pathInfos[i].Name()) < strings.ToLower(pathInfos[j].Name())
		}
	}

	sort.Slice(pathInfos, func(i, j int) bool { return sortFunc(i, j) == sortAsc })
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
		App:          version.App,
		AssetsPrefix: AssetsPrefix,
		Sort: sortInfo{
			Column: sortColumn,
			Asc:    sortAsc,
		},
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
