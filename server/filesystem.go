package server

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FileSystem provides acess to files and directories under a certain root.
// It can optionally optionally:
//
//   - serve .htm(l) files for the corresponding path without suffix, if the
//     original path is not found
//   - hide dotfiles
//   - allow access to file/directories outside the filesystem root via symlinks
type FileSystem struct {
	ResolveHTML          bool
	HideDotFiles         bool
	AllowOutsideSymlinks bool
	Root                 string
}

// Open returns a File object for the specified path under the FileSystem
// directory.
func (fs FileSystem) Open(name string) (*File, error) {
	if fs.HideDotFiles && containsDotFile(name) {
		// Even if the file exists, return 404
		return nil, os.ErrNotExist
	}

	_, err := fs.open(name)
	if os.IsNotExist(err) && fs.ResolveHTML && !(strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm")) {
		for _, suffix := range []string{".html", ".htm"} {
			newName := name + suffix
			if file, err := fs.OpenFile(newName); err == nil {
				return file, nil
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return fs.newFile(name)
}

// OpenFile returns a File object for the specified path under the FileSystem
// directory if it esists and it's not a directory.
func (fs FileSystem) OpenFile(name string) (*File, error) {
	if file, err := fs.open(name); err == nil {
		if fileInfo, err := file.Stat(); err == nil && !fileInfo.IsDir() {
			return fs.newFile(name)
		}
	}
	return nil, os.ErrNotExist
}

func (fs FileSystem) open(name string) (*os.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, fmt.Errorf("invalid character in file path: %s", name)
	}
	fullName := filepath.Join(fs.Root, filepath.FromSlash(path.Clean("/"+name)))
	return os.Open(fullName)
}

func (fs FileSystem) newFile(name string) (*File, error) {
	path, err := fs.resolvePath(filepath.Join(fs.Root, name))
	if err != nil {
		return nil, err
	}
	if !fs.AllowOutsideSymlinks {
		root, err := fs.resolvePath(fs.Root)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, root) {
			return nil, os.ErrPermission
		}
	}
	return NewFile(path, fs.HideDotFiles)
}

func (fs FileSystem) resolvePath(path string) (string, error) {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

// File is an entry of a  FileSystem entry.
//
// If the entry is a directory and the filesystem is configured to hide
// dotfiles, the directory will also not list dotfiles under it.
type File struct {
	Info os.FileInfo

	absPath      string
	hideDotFiles bool
}

// NewFile returns a File for an absolute path.
func NewFile(absPath string, hideDotFiles bool) (*File, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	return &File{
		Info:         info,
		absPath:      absPath,
		hideDotFiles: hideDotFiles,
	}, nil
}

// AbsPath returns the absolute path of the File.
func (f File) AbsPath() string {
	return f.absPath
}

// Readdir files in the directory, excluding special files and optionally
// hidden files (that start with a dot).
func (f File) Readdir() ([]*File, error) {
	file, err := os.Open(f.absPath)
	if err != nil {
		return nil, err
	}
	// don't use Readdir to get the FileInfo since it doesn't resolve
	// symlinks. We want the FileInfo to be the one of the symlink target
	// (hence the use of os.Stat below).
	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	files := make([]*File, 0, len(names))
	for _, name := range names {
		info, err := os.Stat(filepath.Join(f.absPath, name))
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		name := info.Name()
		if f.hideDotFiles && strings.HasPrefix(name, ".") {
			continue
		}
		mode := info.Mode()
		if !(mode.IsDir() || mode.IsRegular()) {
			continue
		}
		file, err := f.newFile(name)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		files = append(files, file)
	}
	return files, nil
}

func (f File) newFile(name string) (*File, error) {
	absPath := filepath.Join(f.AbsPath(), name)
	return NewFile(absPath, f.hideDotFiles)
}

// containsDotFile reports whether name contains a path element starting with a
// period.  The name is assumed to be a delimited by forward slashes, as
// guaranteed by the http.FileSystem interface.
func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}
