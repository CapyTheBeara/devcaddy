package lib

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/fsnotify.v0"
)

func NewFile(name string, op fsnotify.Op) *File {
	f := File{Name: name, Op: op}
	f.GetFile()
	return &f
}

type File struct {
	Name        string
	Content     string
	Type        string
	Dir         string
	Ext         string
	Files       []string
	Error       error
	PluginNames []string `json:"plugins"`
	LogOnly     bool
	Op          fsnotify.Op
}

func (f *File) IsDeleted() bool {
	return f.Op == fsnotify.Remove || f.Op == fsnotify.Rename
}

func (f *File) GetFile() {
	if f.IsDeleted() {
		return
	}

	b, err := ioutil.ReadFile(f.Name)
	if err != nil {
		log.Fatalln(err)
	}

	f.Content = string(b)
}

func (file *File) MergeStoreFiles(s *Store) string {
	contents := []string{}
	dir := filepath.Join(s.Root, file.Dir)

	if len(file.Files) > 0 {
		for _, f := range file.Files {
			path := filepath.Join(dir, f)
			contents = append(contents, s.Get(path))
		}
	} else {
		for _, f := range s.GetAll() {
			if strings.Contains(f.Name, dir) && strings.HasSuffix(f.Name, "."+file.Ext) {
				contents = append(contents, f.Content)
			}
		}
	}

	return strings.Join(contents, "\n")
}

func (f *File) filterFn(dir string) (filter func(string) bool) {
	if len(f.Files) > 0 {
		files := make(map[string]int)

		for _, name := range f.Files {
			files[filepath.Join(dir, name)] = 1
		}

		return func(path string) bool {
			_, ok := files[path]
			return ok
		}
	} else {
		return func(path string) bool {
			return strings.HasSuffix(path, f.Ext)
		}
	}
}
