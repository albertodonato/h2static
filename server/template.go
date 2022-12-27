package server

import (
	_ "embed" // for embed directive
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"sort"
	"strings"

	"github.com/albertodonato/h2static/version"
)

// AssetsPrefix defines the URL prefix for static assets.
const AssetsPrefix = "/.h2static-assets/"

// CSSAsset defines the path of the CSS file.
const CSSAsset = AssetsPrefix + "style.css"

// LogoAsset defines the path of the logo.
const LogoAsset = AssetsPrefix + "logo.svg"

//go:embed template.html
var dirListingTemplateText string

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

type osInfo struct {
	OS   string
	Arch string
}

type templateContext struct {
	App       version.Version
	OS        osInfo
	Dir       DirInfo
	BasePath  string
	CSSAsset  string
	LogoAsset string
	Sort      sortInfo
}

// DirectoryListingTemplateConfig holds configuration for a DirectoryListingTemplate
type DirectoryListingTemplateConfig struct {
	PathPrefix string
}

// DirectoryListingTemplate is a template rendered for a directory.
type DirectoryListingTemplate struct {
	Config   DirectoryListingTemplateConfig
	template *template.Template
}

// NewDirectoryListingTemplate returns a DirectoryListingTemplate for the specified directory.
func NewDirectoryListingTemplate(config DirectoryListingTemplateConfig) *DirectoryListingTemplate {
	return &DirectoryListingTemplate{
		Config: config,
		template: template.Must(
			template.New("DirListing").Funcs(template.FuncMap{
				"inc": func(n int) int { return n + 1 },
			}).Parse(dirListingTemplateText)),
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
	files, err := dir.Readdir()
	if err != nil {
		return nil, err
	}

	var sortFunc func(int, int) bool
	if sortColumn == "s" { // sort by size
		sortFunc = func(i, j int) bool {
			return files[i].Info.Size() < files[j].Info.Size()
		}
	} else { // sort by name
		sortFunc = func(i, j int) bool {
			return strings.ToLower(files[i].Info.Name()) < strings.ToLower(files[j].Info.Name())
		}
	}

	sort.Slice(files, func(i, j int) bool { return sortFunc(i, j) == sortAsc })
	entries := make([]DirEntryInfo, len(files))
	for i, f := range files {
		name := string(template.URL(f.Info.Name()))
		size := f.Info.Size()
		entry := DirEntryInfo{
			Name:  name,
			IsDir: f.Info.IsDir(),
			Size:  size,
		}
		if !f.Info.IsDir() {
			entry.HumanSize = getHumanByteSize(size)
		}
		entries[i] = entry
	}

	return &templateContext{
		App: version.App,
		OS: osInfo{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Dir: DirInfo{
			Name:    path,
			IsRoot:  path == "/",
			Entries: entries,
		},
		BasePath:  t.Config.PathPrefix,
		CSSAsset:  t.Config.PathPrefix + CSSAsset,
		LogoAsset: t.Config.PathPrefix + LogoAsset,
		Sort: sortInfo{
			Column: sortColumn,
			Asc:    sortAsc,
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
